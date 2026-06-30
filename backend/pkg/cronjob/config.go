package cronjob

import "time"

type JobConfig struct {
	Cron       string    `yaml:"cron"`
	Enable     bool      `yaml:"enable"`
	Singleton  bool      `yaml:"singleton"`
	JobTimeout time.Time `yaml:"job_timeout"`
}
