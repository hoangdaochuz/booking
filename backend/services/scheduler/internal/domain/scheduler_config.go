package domain

import (
	"time"

	"github.com/google/uuid"
)

type SchedulerConfig struct {
	Id                 uuid.UUID
	Name               string
	IsEnabled          bool
	IntervalExpression string
	Timeout            time.Duration
	Version            int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// type SchedulerConfigChangedPayload struct {
// 	IsEnable           bool
// 	IntervalExpression string
// }
