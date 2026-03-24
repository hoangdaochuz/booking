package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SagaStatus string

var (
	SAGA_PENDING          SagaStatus = "PENDING"
	SAGA_WAIT_FOR_PAYMENT SagaStatus = "WAIT_FOR_PAYMENT"
	SAGA_PROCESSING       SagaStatus = "PROCESSING"
	SAGA_COMPLETED        SagaStatus = "COMPLETED"
	SAGA_ROLLING_BACK     SagaStatus = "ROLLING_BACK"
	SAGA_ROLLED_BACK      SagaStatus = "ROLLED_BACK"
	SAGA_FAIL             SagaStatus = "FAIL"
)

type Saga struct {
	ID               uuid.UUID
	BookingID        uuid.UUID
	Name             string
	Steps            []SagaStep
	Status           SagaStatus
	CurrentStepIndex int
}

type SagaStepStatus string

var (
	SAGA_STEP_PENDING      SagaStepStatus = "PENDING"
	SAGA_STEP_EXECUTING    SagaStepStatus = "EXECUTING"
	SAGA_STEP_COMPLETED    SagaStepStatus = "COMPLETED"
	SAGA_STEP_COMPENSATING SagaStepStatus = "COMPENSATING"
	SAGA_STEP_COMPENSATED  SagaStepStatus = "COMPENSATED"
	SAGA_STEP_FAILED       SagaStepStatus = "FAILED"
)

type SagaStep struct {
	ID                    uuid.UUID
	SagaID                uuid.UUID
	Name                  string
	Execute               func(ctx context.Context) error
	Compensate            func(ctx context.Context) error
	ExecutedAt            time.Time
	CompenstatedAt        time.Time
	Status                SagaStepStatus
	Order                 int16
	ShouldPauseForPayment bool
}
