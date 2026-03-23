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
}

type CreateSagaResponse struct{}

func (s *SagaService) CreateSaga(ctx context.Context, req *CreateSagaRequest) (*CreateSagaResponse, error) {
	return nil, nil
}

func (s *SagaService) UpdateSaga(ctx context.Context, req *CreateSagaRequest) error {
	return nil
}

func (s *SagaService) GetSagaById(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	return nil, nil
}

func (s *SagaService) GetSagaByBookingId(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	return nil, nil
}

func (s *SagaService) HandleSagaAferPaymentSuccess(ctx context.Context, req json.RawMessage) error {
	return nil
}

func (s *SagaService) HandleSagaAfterPaymentFailure(ctx context.Context, req json.RawMessage) error {
	return nil
}
