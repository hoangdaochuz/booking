package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
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

type SagaStepProcessor struct {
	Execute    *domain.SagaStepFunc
	Compensate *domain.SagaStepFunc
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
	saga            *domain.Saga
	sagaRepo        repository.SagaRepositoryInterface
	logger          *zap.Logger
	paymentResponse *paymentv1.CreatePaymentResponse
}

func NewSagaHandler(saga *domain.Saga, sagaRepo repository.SagaRepositoryInterface, logger *zap.Logger) *SagaHandler {
	return &SagaHandler{
		saga:     saga,
		logger:   logger,
		sagaRepo: sagaRepo,
	}
}

func (s *SagaHandler) GetSaga() *domain.Saga {
	return s.saga
}

func (s *SagaHandler) AddStep(step *domain.SagaStep) {
	s.saga.Steps = append(s.saga.Steps, *step)
}

func (s *SagaHandler) GetSagaID() uuid.UUID {
	return s.saga.ID
}

func (s *SagaHandler) GetPaymentResponse() *paymentv1.CreatePaymentResponse {
	return s.paymentResponse
}

func (s *SagaHandler) SetPaymentResponse(res *paymentv1.CreatePaymentResponse) {
	s.paymentResponse = res
}

func (s *SagaHandler) Execute(ctx context.Context, index int) error {
	s.logger.Sugar().Infof("Executing saga...")
	s.saga.Status = domain.SAGA_PROCESSING
	for index < len(s.saga.Steps) {
		step := s.saga.Steps[index]
		s.saga.Steps[index].Status = domain.SAGA_STEP_EXECUTING
		s.logger.Sugar().Infof("Executing saga step %s", step.Name)
		if err := step.Execute(ctx); err != nil {
			s.logger.Sugar().Errorf("[Saga - %s]: Step %d - %s execute fail", s.saga.Name, step.Order, step.Name, zap.Error(err))
			s.saga.Status = domain.SAGA_ROLLING_BACK
			s.saga.Steps[index].Status = domain.SAGA_STEP_FAILED
			err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
			if err != nil {
				return err
			}
			return s.compensate(ctx, index)
		}
		s.saga.Steps[index].Status = domain.SAGA_STEP_COMPLETED
		s.saga.Steps[index].ExecutedAt = time.Now()
		if s.saga.Steps[index].ShouldPauseForPayment {
			s.saga.Status = domain.SAGA_WAIT_FOR_PAYMENT
			// s.saga.CurrentStepIndex = s.saga.CurrentStepIndex + 1
			err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
			if err != nil {
				return err
			}
			s.logger.Sugar().Infof("Saga step completed and Saga exist to wait for payment successfully...")
			return nil
		}
		index++
		s.saga.CurrentStepIndex = index
		err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
		if err != nil {
			return err
		}
		s.logger.Sugar().Infof("Saga step completed successfully...")
	}
	s.saga.Status = domain.SAGA_COMPLETED
	err := s.sagaRepo.UpdateSaga(ctx, s.saga, nil)
	if err != nil {
		return err
	}
	s.logger.Sugar().Infof("Saga completed successfully")
	return nil
}

func (s *SagaHandler) compensate(ctx context.Context, index int) error {

	isRolledback := true
	s.logger.Sugar().Infof("Compensating saga")
	for i := index - 1; i >= 0; i-- {
		s.saga.Steps[i].Status = domain.SAGA_STEP_COMPENSATING
		step := s.saga.Steps[i]
		s.logger.Sugar().Infof("Compensating step %s ", step.Name)
		if err := step.Compensate(ctx); err != nil {
			s.saga.Steps[i].Status = domain.SAGA_STEP_FAILED
			isRolledback = false
			err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
			if err != nil {
				return err
			}
			s.logger.Sugar().Error("[Saga - %s]: Step %d - %s compensate fail", s.saga.Name, step.Order, s.saga.Steps[i].Name, zap.Error(err))
			continue
		}
		s.saga.Steps[i].Status = domain.SAGA_STEP_COMPENSATED
		s.saga.Steps[i].CompenstatedAt = time.Now()
		s.saga.CurrentStepIndex = i
		err := s.sagaRepo.UpdateSaga(ctx, s.saga, &step.ID)
		if err != nil {
			return err
		}
		s.logger.Sugar().Infof("Compensated step %s successfully", step.Name)
	}
	if isRolledback {
		s.saga.Status = domain.SAGA_ROLLED_BACK
	} else {
		s.saga.Status = domain.SAGA_FAIL
	}
	err := s.sagaRepo.UpdateSaga(ctx, s.saga, nil)
	if err != nil {
		return err
	}
	s.logger.Sugar().Infof("Compensated saga successfully")
	return nil
}
