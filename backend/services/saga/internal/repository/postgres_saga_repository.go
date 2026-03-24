package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	query := `INSERT INTO sagas(id, booking_id, name, status, current_step_index, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.Exec(ctx, query, saga.ID, saga.BookingID, saga.Name, string(saga.Status), saga.CurrentStepIndex, time.Now())
	if err != nil {
		return fmt.Errorf("Insert sagas fail: %w", err)
	}

	for _, step := range saga.Steps {
		query = `INSERT INTO saga_steps(id , saga_id, name, status, order, should_pause_for_payment)
				 VALUES($1, $2, $3, $4, $5, $6)`
		_, err = tx.Exec(ctx, query, step.ID, step.SagaID, step.Name, string(step.Status), step.Order, step.ShouldPauseForPayment)
		if err != nil {
			return fmt.Errorf("Fail to insert into saga_steps: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (s *SagaRepository) GetSagaById(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	saga := domain.Saga{}
	query := `SELECT s.id, s.booking_id, s.name, s.status, s.current_step_index
			 FROM sagas as s WHERE s.id = $1`
	var sagaStatus string
	err := s.pool.QueryRow(ctx, query, id).Scan(&saga.ID, &saga.BookingID, &saga.Name, &sagaStatus, &saga.CurrentStepIndex)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get saga: %w", err)
	}
	saga.Status = domain.SagaStatus(sagaStatus)

	subQuery := `SELECT id, saga_id, name, executed_at, compensated_at, status, order, should_pause_for_payment
				FROM saga_steps WHERE id = $1`

	rows, err := s.pool.Query(ctx, subQuery, saga.ID)
	if err != nil {
		return nil, fmt.Errorf("get saga steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		sagaStep := domain.SagaStep{}
		var sagaStepStatus string
		if err := rows.Scan(&sagaStep.ID, &sagaStep.SagaID, &sagaStep.Name, &sagaStep.ExecutedAt, &sagaStep.CompenstatedAt, &sagaStepStatus, &sagaStep.Order, &sagaStep.ShouldPauseForPayment); err != nil {
			return nil, fmt.Errorf("scan saga step: %w", err)
		}
		sagaStep.Status = domain.SagaStepStatus(sagaStepStatus)
		saga.Steps = append(saga.Steps, sagaStep)
	}
	return &saga, nil
}

func (s *SagaRepository) UpdateSaga(ctx context.Context, saga *domain.Saga, stepIdPtr *uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("fail to start transaction for update saga: %w", err)
	}
	defer tx.Rollback(ctx)
	query := `UPDATE sagas SET name = $2, status = $3, current_step_index = $4 WHERE id = $1`
	_, err = tx.Exec(ctx, query, saga.ID, saga.Name, string(saga.Status), saga.CurrentStepIndex)
	if err != nil {
		return fmt.Errorf("Fail to update saga: %w", err)
	}
	if stepIdPtr != nil {
		stepId := *stepIdPtr
		var step domain.SagaStep
		for _, st := range saga.Steps {
			if st.ID.String() == stepId.String() {
				step = st
				break
			}
		}

		query := `UPDATE saga_steps SET executed_at = $2, compensated_at = $3, status = $4 WHERE id = $1`
		_, err = tx.Exec(ctx, query, stepId, step.ExecutedAt, step.CompenstatedAt, string(step.Status))
		if err != nil {
			return fmt.Errorf("Fail to update step saga: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *SagaRepository) GetSagaByBookingId(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	saga := domain.Saga{}
	query := `SELECT s.id, s.booking_id, s.name, s.status, s.current_step_index
			 FROM sagas as s WHERE s.booking_id = $1`
	var sagaStatus string
	err := s.pool.QueryRow(ctx, query, id).Scan(&saga.ID, &saga.BookingID, &saga.Name, &sagaStatus, &saga.CurrentStepIndex)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get saga: %w", err)
	}
	saga.Status = domain.SagaStatus(sagaStatus)

	subQuery := `SELECT id, saga_id, name, executed_at, compensated_at, status, order, should_pause_for_payment
				FROM saga_steps WHERE id = $1`

	rows, err := s.pool.Query(ctx, subQuery, saga.ID)
	if err != nil {
		return nil, fmt.Errorf("get saga steps: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		sagaStep := domain.SagaStep{}
		var sagaStepStatus string
		if err := rows.Scan(&sagaStep.ID, &sagaStep.SagaID, &sagaStep.Name, &sagaStep.ExecutedAt, &sagaStep.CompenstatedAt, &sagaStepStatus, &sagaStep.Order, &sagaStep.ShouldPauseForPayment); err != nil {
			return nil, fmt.Errorf("scan saga step: %w", err)
		}
		sagaStep.Status = domain.SagaStepStatus(sagaStepStatus)
		saga.Steps = append(saga.Steps, sagaStep)
	}
	return &saga, nil
}
