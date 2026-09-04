package cronjob

import (
	"context"
	"fmt"
	"sync"
	"time"

	cronv3 "github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/ticketbox/scheduler/internal/domain"
	"github.com/ticketbox/scheduler/internal/repository"
)

// defaultJobTimeout is used when a job config carries no positive timeout,
// otherwise context.WithTimeout(0) would expire immediately.
const defaultJobTimeout = 30 * time.Second

type CronJobManager struct {
	logger         *zap.Logger
	cron           *cronv3.Cron
	jobs           map[string]Job
	jobEntryID     map[string]cronv3.EntryID
	jobConfig      map[string]domain.SchedulerConfig
	mu             sync.Mutex
	isStarted      bool
	withSecond     bool
	withErrHandler func(ctx context.Context, err error) error
	schedulerRepo  repository.SchedulerRepository
}

func NewCronjobManager(logger *zap.Logger, schedulerRepo repository.SchedulerRepository, opts ...Option) *CronJobManager {
	cronjobManager := &CronJobManager{
		jobs:          make(map[string]Job),
		jobConfig:     make(map[string]domain.SchedulerConfig),
		mu:            sync.Mutex{},
		jobEntryID:    make(map[string]cronv3.EntryID),
		logger:        logger,
		schedulerRepo: schedulerRepo,
	}

	for _, opt := range opts {
		opt(cronjobManager)
	}

	cronOpts := []cronv3.Option{}

	if cronjobManager.withSecond {
		cronOpts = append(cronOpts, cronv3.WithSeconds())
	}

	cronjobManager.cron = cronv3.New(cronOpts...)

	return cronjobManager

}

// Start is non-blocking (robfig cron spawns its own goroutine), so callers
// must NOT wrap it in an extra `go func()`. It is idempotent and safe to
// call once after all jobs are registered.
func (c *CronJobManager) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isStarted {
		return
	}
	c.cron.Start()
	c.isStarted = true
	c.logger.Info("[CronJobManager][Start] Cronjob manager starts successfully")

}

func (c *CronJobManager) LoadJobSchedulerConfigs(ctx context.Context) error {
	jobConfigs, err := c.schedulerRepo.ListSchedulersConfig(ctx)
	if err != nil {
		return fmt.Errorf("[CronJobManager][LoadJobSchedulerConfigs] Load job schedule config fail: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jobConfig = make(map[string]domain.SchedulerConfig, len(jobConfigs))
	for _, jc := range jobConfigs {
		c.jobConfig[jc.Name] = jc
	}
	c.logger.Info("[CronJobManager][LoadJobSchedulerConfigs] Scheduler configs load successfully",
		zap.Int("count", len(jobConfigs)))
	return nil
}

func (c *CronJobManager) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isStarted {
		return nil
	}
	c.isStarted = false
	return c.cron.Stop().Err()
}

func (c *CronJobManager) RegisterJob(_ context.Context, job Job) error {
	jobName := job.Name()

	c.mu.Lock()
	jc, ok := c.jobConfig[jobName]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("[CronJobManager][RegisterJob] Job config of this job is not exist: %s", jobName)
	}
	c.jobs[jobName] = job
	if !jc.IsEnabled {
		c.mu.Unlock()
		c.logger.Info("[CronJobManager][RegisterJob] Job is disabled, skip scheduling",
			zap.String("job", jobName))
		return nil
	}
	handler := jobHandle(job, c.logger, jc.Timeout, c.withErrHandler)
	entryID, err := c.cron.AddFunc(jc.IntervalExpression, handler)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("[CronJobManager][RegisterJob] Register job fail: %w", err)
	}
	c.jobEntryID[jobName] = entryID
	c.mu.Unlock()

	c.logger.Info("[CronJobManager][RegisterJob] Job registered successfully",
		zap.String("job", jobName),
		zap.String("cron", jc.IntervalExpression),
		zap.Duration("timeout", jc.Timeout),
		zap.Bool("isEnabled", jc.IsEnabled))
	return nil
}

func jobHandle(job Job, logger *zap.Logger, timeout time.Duration, handlerErr func(ctx context.Context, err error) error) func() {
	if timeout <= 0 {
		timeout = defaultJobTimeout
	}
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := job.Run(ctx); err != nil {
			logger.Error("Job run failed",
				zap.String("job", job.Name()),
				zap.Error(err))
			if handlerErr != nil {
				if handleErr := handlerErr(ctx, err); handleErr != nil {
					logger.Error("Job error handler failed",
						zap.String("job", job.Name()),
						zap.Error(handleErr))
				}
			}
			return
		}
		logger.Info("Job completed successfully", zap.String("job", job.Name()))
	}
}

func (c *CronJobManager) ReRegisterJob(_ context.Context, jc domain.SchedulerConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentJobConfig, ok := c.jobConfig[jc.Name]
	if !ok {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] Fail to get job config: %s", jc.Name)
	}
	if jc.Version <= currentJobConfig.Version {
		// NOOP: stale event, keep the current schedule untouched.
		c.logger.Info("[CronJobManager][ReRegisterJob] The target job config is not latest version, skip",
			zap.String("job", jc.Name),
			zap.Int("currentVersion", currentJobConfig.Version),
			zap.Int("targetVersion", jc.Version))
		return nil
	}

	// Remove the old schedule if there is one (a disabled job has none).
	if jobEntryID, ok := c.jobEntryID[jc.Name]; ok {
		c.cron.Remove(jobEntryID)
		delete(c.jobEntryID, jc.Name)
	}

	c.jobConfig[jc.Name] = jc

	if !jc.IsEnabled {
		c.logger.Info("[CronJobManager][ReRegisterJob] Job is disabled, schedule removed",
			zap.String("job", jc.Name))
		return nil
	}

	job, ok := c.jobs[jc.Name]
	if !ok {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] Fail to get job implementation: %s", jc.Name)
	}
	entryID, err := c.cron.AddFunc(jc.IntervalExpression, jobHandle(job, c.logger, jc.Timeout, c.withErrHandler))
	if err != nil {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] fail to register job: %w", err)
	}
	c.jobEntryID[jc.Name] = entryID
	c.logger.Info("[CronJobManager][ReRegisterJob] Job re-registered successfully",
		zap.String("job", jc.Name),
		zap.String("cron", jc.IntervalExpression),
		zap.Duration("timeout", jc.Timeout),
		zap.Bool("isEnabled", jc.IsEnabled))
	return nil
}
