package cronjob

import (
	"context"
	"sync"

	cronv3 "github.com/robfig/cron/v3"
)

type CronJobManager struct {
	cron           *cronv3.Cron
	jobs           map[string]Job
	jobConfig      map[string]JobConfig
	mu             sync.Mutex
	isStarted      bool
	withSecond     bool
	withErrHandler func(ctx context.Context) error
}

func New(opts ...Option) *CronJobManager {
	cronjobManager := &CronJobManager{
		jobs:      make(map[string]Job),
		jobConfig: make(map[string]JobConfig),
		mu:        sync.Mutex{},
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

func (c *CronJobManager) Start(ctx context.Context) error {
	return nil
}

func (c *CronJobManager) Stop(ctx context.Context) error {
	return nil
}

func (c *CronJobManager) RegisterJob(job Job) error {
	return nil
}
