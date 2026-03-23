package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ticketbox/saga/internal/domain"
)

type SagaRepository struct {
	pool *pgxpool.Pool
}

func NewSagaRepository(pool *pgxpool.Pool) *SagaRepository {
	return &SagaRepository{
		pool: pool,
	}
}

func (s *SagaRepository) Create(ctx context.Context, saga *domain.Saga) error {
	return nil
}

func (s *SagaRepository) GetSagaById(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	return nil, nil
}

func (s *SagaRepository) UpdateSaga(ctx context.Context, saga *domain.Saga) error {
	return nil
}

func (s *SagaRepository) GetSagaByBookingId(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	return nil, nil
}
