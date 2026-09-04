package service_outbound

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ticketbox/pkg/outbound"
	"github.com/ticketbox/scheduler/internal/kafka"
	"github.com/ticketbox/scheduler/internal/repository"
	"go.uber.org/zap"
)

type OutboundEventService struct {
	logger                *zap.Logger
	repo                  repository.OutboundEventRepository
	outboundEventProducer *kafka.OutboundEventProducer
}

const (
	PendingOutboundEventLimit = 50
)

func NewOutboundEventService(logger *zap.Logger, repo repository.OutboundEventRepository, producer *kafka.OutboundEventProducer) *OutboundEventService {
	return &OutboundEventService{
		logger:                logger,
		repo:                  repo,
		outboundEventProducer: producer,
	}
}

func (o *OutboundEventService) ListPendingOutboundEvents(ctx context.Context) ([]outbound.OutboundEvent, error) {
	offset := 0
	isExist := true
	pendingOutboundEvents := []outbound.OutboundEvent{}
	for isExist {
		events, err := o.repo.GetListOutboundPending(ctx, PendingOutboundEventLimit, offset)
		if err != nil {
			o.logger.Sugar().Error("[OutboundEventService][ListPendingOutboundEvents] Fail to get list pending outbound events", zap.Error(err))
			return pendingOutboundEvents, fmt.Errorf("Fail to get list pending outbound events: %w", err)
		}
		if len(events) < PendingOutboundEventLimit {
			isExist = false
		} else {
			offset += PendingOutboundEventLimit
		}
		pendingOutboundEvents = append(pendingOutboundEvents, events...)
	}
	return pendingOutboundEvents, nil
}

func (o *OutboundEventService) MarkOutboundEventAsPublished(ctx context.Context, id uuid.UUID) error {
	return o.repo.MarkEventAsPublished(ctx, id)
}

func (o *OutboundEventService) MarkOutboundEventsAsPublished(ctx context.Context, ids []uuid.UUID) error {
	return o.repo.MarkEventsAsPublished(ctx, ids)
}
