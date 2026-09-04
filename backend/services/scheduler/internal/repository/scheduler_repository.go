package repository

import (
	"context"
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

func (s *SchedulerConfigRepo) UpdateById(ctx context.Context, id uuid.UUID, target domain.SchedulerConfig) error {
	query := `UPDATE scheduler_configs SET is_enable = $2, interval_expression = $3, updated_at = now(), version = version + 1 WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id, target.IsEnabled, target.IntervalExpression)
	if err != nil {
		return err
	}
	return nil
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
		cfg.Timeout = time.Duration(timeout * int32(time.Second))
		schedulerCfgs = append(schedulerCfgs, cfg)
	}
	return schedulerCfgs, nil
}
