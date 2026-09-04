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

func (c *CronJobManager) Start(ctx context.Context) {
	c.cron.Start()
	c.isStarted = true
	c.logger.Sugar().Info("[CronJobManager][Start] Cronjob manager starts successfully")

}

func (c *CronJobManager) LoadJobSchedulerConfigs(ctx context.Context) error {
	c.jobConfig = make(map[string]domain.SchedulerConfig)
	jobConfigs, err := c.schedulerRepo.ListSchedulersConfig(ctx)
	if err != nil {
		return fmt.Errorf("[CronJobManager][LoadJobSchedulerConfigs] Load job schedule config fail: %w", err)
	}
	for _, jc := range jobConfigs {
		c.jobConfig[jc.Name] = jc
	}
	c.logger.Sugar().Info("[CronJobManager][LoadJobSchedulerConfigs] Scheduler configs load successfully")
	return nil
}

func (c *CronJobManager) Stop() error {
	return c.cron.Stop().Err()
}

func (c *CronJobManager) RegisterJob(ctx context.Context, job Job) error {
	jobName := job.Name()
	jc, ok := c.jobConfig[jobName]
	if !ok {
		return fmt.Errorf("[CronJobManager][RegisterJob] Job config of this job is not exist")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jobs[jobName] = job

	entryId, err := c.cron.AddFunc(jc.IntervalExpression, jobHandle(job, c.logger, jc.Timeout, c.withErrHandler))
	if err != nil {
		return fmt.Errorf("[CronJobManager][RegisterJob] Register job fail: %w", err)
	}
	c.jobEntryID[jobName] = entryId
	c.logger.Sugar().Infof("[CronJobManager][RegisterJob]: Job %s registered successfully", jobName, zap.String("cron", jc.IntervalExpression), zap.String("timeout", jc.Timeout.String()), zap.Bool("IsEnable", jc.IsEnabled))
	return nil
}

func jobHandle(job Job, logger *zap.Logger, timeout time.Duration, handlerErr func(ctx context.Context, err error) error) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := job.Run(ctx)
		if err != nil {
			handleerr := handlerErr(ctx, err)
			logger.Sugar().Errorf("[%s] handlerErr fail", job.Name(), zap.Error(handleerr))
		} else {
			logger.Sugar().Infof("[%s] Complete successfully", job.Name())
		}
	}
}

func (c *CronJobManager) ReRegisterJob(ctx context.Context, jc domain.SchedulerConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	jobEntryId, ok := c.jobEntryID[jc.Name]
	if !ok {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] Fail to get job entry id")
	}
	c.cron.Remove(jobEntryId)
	currentJobConfig, ok := c.jobConfig[jc.Name]
	if !ok {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] Fail to get job config")
	}
	if jc.Version <= currentJobConfig.Version {
		// NOOP
		c.logger.Sugar().Info("[CronJobManager][ReRegisterJob] The target job config is not latest version")
	}

	delete(c.jobConfig, jc.Name)
	c.jobConfig[jc.Name] = jc

	if !jc.IsEnabled {
		return nil
	}

	job, ok := c.jobs[jc.Name]
	if !ok {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] Fail to get job implementation")
	}
	entryId, err := c.cron.AddFunc(jc.IntervalExpression, jobHandle(job, c.logger, jc.Timeout, c.withErrHandler))
	if err != nil {
		return fmt.Errorf("[CronJobManager][ReRegisterJob] fail to register job: %w", err)
	}
	c.jobEntryID[jc.Name] = entryId
	c.logger.Sugar().Infof("[CronJobManager][RegisterJob]: Job %s re-registered successfully", jc.Name, zap.String("cron", jc.IntervalExpression), zap.String("timeout", jc.Timeout.String()), zap.Bool("IsEnable", jc.IsEnabled))
	return nil
}
