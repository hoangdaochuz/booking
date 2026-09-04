package cronjob

import "context"

type Option func(c *CronJobManager)

func WithSecond(c *CronJobManager) {
	c.withSecond = true
}

func WithCronjobErrHandler(c *CronJobManager) {
	c.withErrHandler = func(ctx context.Context, err error) error {
		// TODO
		return nil
	}
}
