package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/ticketbox/saga/internal/domain"
	"github.com/ticketbox/saga/internal/repository"
	"go.uber.org/zap"
)

type SagaRegistry struct {
	sagas  map[string]*domain.Saga
	logger *zap.Logger
}

func NewSagaRegistry() *SagaRegistry {
	return &SagaRegistry{
		sagas: make(map[string]*domain.Saga),
	}
}

func (s *SagaRegistry) Register(saga *domain.Saga) error {
	if _, ok := s.sagas[saga.ID.String()]; !ok {
		return fmt.Errorf("Saga already registered")
	}
	s.sagas[saga.ID.String()] = saga
	return nil
}

type SagaStepFunc func(ctx context.Context) error

type SagaStepProcessor struct {
	Execute    SagaStepFunc
	Compensate SagaStepFunc
}

type SagaStepRegistry struct {
	store map[string]SagaStepProcessor
}

func NewSagaStepRegistry() *SagaStepRegistry {
	return &SagaStepRegistry{
		store: make(map[string]SagaStepProcessor),
	}
}

func (s *SagaStepRegistry) Register(name string, processor SagaStepProcessor) {
	if _, ok := s.store[name]; ok {
		return
	}
	s.store[name] = processor
}

type SagaHandler struct {
	saga     *domain.Saga
	sagaRepo repository.SagaRepositoryInterface
	logger   *zap.Logger
}

func NewSagaHandler(saga *domain.Saga, logger *zap.Logger) *SagaHandler {
	return &SagaHandler{
		saga:   saga,
		logger: logger,
	}
}

func (s *SagaHandler) AddStep(step *domain.SagaStep) {
	s.saga.Steps = append(s.saga.Steps, *step)
}

func (s *SagaHandler) Execute(ctx context.Context, index int) (*domain.Saga, error) {
	s.saga.Status = domain.SAGA_PROCESSING
	for index < len(s.saga.Steps) {
		step := s.saga.Steps[index]
		s.saga.Steps[index].Status = domain.SAGA_STEP_EXECUTING
		if err := step.Execute(ctx); err != nil {
			s.logger.Sugar().Errorf("[Saga - %s]: Step %d - %s execute fail", s.saga.Name, step.Order, step.Name, zap.Error(err))
			s.saga.Status = domain.SAGA_ROLLING_BACK
			s.saga.Steps[index].Status = domain.SAGA_STEP_FAILED
			err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
			if err != nil {
				return nil, err
			}
			return s.compensate(ctx, index)
		}
		s.saga.Steps[index].Status = domain.SAGA_STEP_COMPLETED
		s.saga.Steps[index].ExecutedAt = time.Now()
		if s.saga.Steps[index].ShouldPauseForPayment {
			s.saga.Status = domain.SAGA_PENDING
			// s.saga.CurrentStepIndex = s.saga.CurrentStepIndex + 1
			err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
			if err != nil {
				return nil, err
			}
			return s.saga, nil
		}
		index++
		s.saga.CurrentStepIndex = index
		err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
		if err != nil {
			return nil, err
		}
	}
	s.saga.Status = domain.SAGA_COMPLETED
	err := s.sagaRepo.UpdateSaga(ctx, s.saga, nil)
	if err != nil {
		return nil, err
	}
	return s.saga, nil
}

func (s *SagaHandler) compensate(ctx context.Context, index int) (*domain.Saga, error) {
	isRolledback := true
	for i := index - 1; i >= 0; i-- {
		s.saga.Steps[i].Status = domain.SAGA_STEP_COMPENSATING
		step := s.saga.Steps[i]
		if err := step.Compensate(ctx); err != nil {
			s.saga.Steps[i].Status = domain.SAGA_STEP_FAILED
			isRolledback = false
			err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
			if err != nil {
				return nil, err
			}
			s.logger.Sugar().Error("[Saga - %s]: Step %d - %s compensate fail", s.saga.Name, step.Order, s.saga.Steps[i].Name, zap.Error(err))
			continue
		}
		s.saga.Steps[i].Status = domain.SAGA_STEP_COMPENSATED
		s.saga.Steps[i].CompenstatedAt = time.Now()
		s.saga.CurrentStepIndex = i
		err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
		if err != nil {
			return nil, err
		}
	}
	if isRolledback {
		s.saga.Status = domain.SAGA_ROLLED_BACK
	} else {
		s.saga.Status = domain.SAGA_FAIL
	}
	err := s.sagaRepo.UpdateSaga(ctx, s.saga, nil)
	if err != nil {
		return nil, err
	}
	return s.saga, nil
}
