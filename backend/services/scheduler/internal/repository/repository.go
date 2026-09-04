package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ticketbox/pkg/outbound"
	"github.com/ticketbox/scheduler/internal/domain"
)

type SchedulerRepository interface {
	ListSchedulersConfig(ctx context.Context) ([]domain.SchedulerConfig, error)
	UpdateById(ctx context.Context, id uuid.UUID, target domain.SchedulerConfig) error
}

type OutboundEventRepository interface {
	Create(ctx context.Context, outboundEvent outbound.OutboundEvent) error
	GetListOutboundPending(ctx context.Context, limit, offset int) ([]outbound.OutboundEvent, error)
	MarkEventAsPublished(ctx context.Context, id uuid.UUID) error
	MarkEventsAsPublished(ctx context.Context, ids []uuid.UUID) error
}
