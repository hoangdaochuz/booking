package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ticketbox/saga/internal/domain"
	"github.com/ticketbox/saga/internal/registry"
	"github.com/ticketbox/saga/internal/repository"
	"go.uber.org/zap"
)

type SagaService struct {
	logger           *zap.Logger
	repo             repository.SagaRepositoryInterface
	sagaStepRegistry *registry.SagaStepRegistry
}

func NewSagaService(logger *zap.Logger, repo repository.SagaRepositoryInterface) *SagaService {
	return &SagaService{
		logger: logger,
		repo:   repo,
	}
}

type CreateSagaRequest struct {
	Name      string
	BookingId uuid.UUID
	Steps     []domain.SagaStep
}

type UpdateSagaRequest struct {
	Id               uuid.UUID
	Name             string
	Status           domain.SagaStatus
	CurrentStepIndex int
}

func (s *SagaService) CreateSaga(ctx context.Context, req *CreateSagaRequest) (*domain.Saga, error) {
	sagaCreate := domain.Saga{
		ID:               uuid.New(),
		BookingID:        req.BookingId,
		Name:             req.Name,
		Steps:            req.Steps,
		Status:           domain.SAGA_PENDING,
		CurrentStepIndex: 0,
	}
	err := s.repo.Create(ctx, &sagaCreate)
	if err != nil {
		return nil, err
	}
	return &sagaCreate, nil
}

func (s *SagaService) UpdateSaga(ctx context.Context, req *UpdateSagaRequest) error {
	sagaUpdate := domain.Saga{
		ID:               req.Id,
		Name:             req.Name,
		Status:           req.Status,
		CurrentStepIndex: req.CurrentStepIndex,
	}
	return s.repo.UpdateSaga(ctx, &sagaUpdate, nil)
}

func (s *SagaService) GetSagaById(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {

	return s.repo.GetSagaById(ctx, id)
}

func (s *SagaService) GetSagaByBookingId(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	return s.repo.GetSagaByBookingId(ctx, id)
}

func (s *SagaService) StartOrderSaga(ctx context.Context) error {
	return nil
}

func (s *SagaService) HandleSagaAferPaymentSuccess(ctx context.Context, req json.RawMessage) error {
	// TODO
	s.logger.Info("Handling Saga after payment process successfully")
	return nil
}

func (s *SagaService) HandleSagaAfterPaymentFailure(ctx context.Context, req json.RawMessage) error {
	// TODO
	s.logger.Info("Handling Saga after payment process fail")
	return nil
}
