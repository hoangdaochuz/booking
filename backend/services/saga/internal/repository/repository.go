package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ticketbox/saga/internal/domain"
)

var ErrNotFound = errors.New("record not found")

type SagaRepositoryInterface interface {
	Create(ctx context.Context, saga *domain.Saga) error
	GetSagaById(ctx context.Context, id uuid.UUID) (*domain.Saga, error)
	GetSagaByBookingId(ctx context.Context, id uuid.UUID) (*domain.Saga, error)
	UpdateSaga(ctx context.Context, saga *domain.Saga, sagaStepId *uuid.UUID) error
}
