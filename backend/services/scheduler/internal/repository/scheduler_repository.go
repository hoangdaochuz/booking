package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ticketbox/scheduler/internal/domain"
)

type SchedulerConfigRepo struct {
	pool *pgxpool.Pool
	tx   *pgx.Tx
}

func NewSchedulerConfigRepo(pool *pgxpool.Pool, tx *pgx.Tx) *SchedulerConfigRepo {
	return &SchedulerConfigRepo{
		pool: pool,
		tx:   tx,
	}
}

func (s *SchedulerConfigRepo) GetById(ctx context.Context, id uuid.UUID) (*domain.SchedulerConfig, error) {
	return nil, nil
}

func (s *SchedulerConfigRepo) UpdateById(ctx context.Context, id uuid.UUID, target domain.SchedulerConfig) (*domain.SchedulerConfig, error) {
	query := `UPDATE scheduler_configs SET is_enable = $2, interval_expression = $3, timeout = $4, updated_at = now(), version = version + 1 WHERE id = $1 RETURNING id, name, timeout, version, interval_expression, is_enable, created_at, updated_at`
	row := s.pool.QueryRow(ctx, query, id, target.IsEnabled, target.IntervalExpression, int32(target.Timeout/time.Second))
	var cfg domain.SchedulerConfig
	var timeout int32
	err := row.Scan(&cfg.Id, &cfg.Name, &timeout, &cfg.Version, &cfg.IntervalExpression, &cfg.IsEnabled, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("scheduler config not found")
		}
		return nil, err
	}
	cfg.Timeout = time.Duration(timeout) * time.Second
	return &cfg, nil
}

func (s *SchedulerConfigRepo) ListSchedulersConfig(ctx context.Context) ([]domain.SchedulerConfig, error) {
	query := `SELECT id, name, timeout, version, interval_expression, is_enable, created_at, updated_at FROM scheduler_configs`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	schedulerCfgs := []domain.SchedulerConfig{}
	var cfg domain.SchedulerConfig
	var timeout int32
	for rows.Next() {
		err = rows.Scan(&cfg.Id, &cfg.Name, &timeout, &cfg.Version, &cfg.IntervalExpression, &cfg.IsEnabled, &cfg.CreatedAt, &cfg.UpdatedAt)
		if err != nil {
			continue
		}
		// Convert before multiplying: timeout * int32(time.Second) overflows
		// int32 for any timeout > 2s, wrapping the duration negative.
		cfg.Timeout = time.Duration(timeout) * time.Second
		schedulerCfgs = append(schedulerCfgs, cfg)
	}
	return schedulerCfgs, nil
}
