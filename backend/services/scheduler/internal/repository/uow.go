package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWorkRepository interface {
	OutboundRepo() OutboundEventRepository
	SchedulerConfigRepo() SchedulerRepository
}

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(repos *UowRepos) error) error
}

type UnitOfWorkImpl struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWorkImpl {
	return &UnitOfWorkImpl{
		pool: pool,
	}
}

func (u *UnitOfWorkImpl) Execute(ctx context.Context, fn func(repos *UowRepos) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil {
			fmt.Println("fail to roll back the transaction: %w", err)
		}
	}()

	repos := &UowRepos{
		schedulerConfigRepo: NewSchedulerConfigRepo(u.pool, &tx),
		outboundEventRepo:   NewOutboundEventRepository(u.pool, &tx),
	}

	err = fn(repos)
	if err != nil {
		return tx.Rollback(ctx)
	}
	return tx.Commit(ctx)
}

type UowRepos struct {
	schedulerConfigRepo *SchedulerConfigRepo
	outboundEventRepo   *OutBoundEventRepo
}

func (u *UowRepos) OutboundRepo() OutboundEventRepository {
	return u.outboundEventRepo
}

func (u *UowRepos) SchedulerConfigRepo() SchedulerRepository {
	return u.schedulerConfigRepo
}
