package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ticketbox/pkg/kafka"
	"go.uber.org/zap"
)

type OutboundCfg struct {
	Interval string
}

type OutboundHandler struct {
	updaterFunc func(ctx context.Context, eventIds []uuid.UUID) error
	loader      func(ctx context.Context) ([]OutboundEvent, error)
	logger      *zap.Logger
	producer    *kafka.Producer
	cfg         OutboundCfg
}

func NewOutboundHandler(logger *zap.Logger, producer *kafka.Producer, cfg OutboundCfg, updaterFunc func(ctx context.Context, eventIds []uuid.UUID) error, loader func(ctx context.Context) ([]OutboundEvent, error)) *OutboundHandler {
	return &OutboundHandler{
		logger:      logger,
		producer:    producer,
		cfg:         cfg,
		updaterFunc: updaterFunc,
		loader:      loader,
	}
}

func (o *OutboundHandler) Start(ctx context.Context) error {
	duration, err := time.ParseDuration(o.cfg.Interval)
	if err != nil {
		o.logger.Sugar().Errorf("[OutboundHandler] Fail to parse time interval expression", zap.Error(err))
		return fmt.Errorf("[OutboundHandler] fail to parse time interval expression: %w", err)
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	o.logger.Sugar().Infoln("[OutboundHandler] Outbound handler is starting")
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					o.logger.Sugar().Errorf("[OutboundHandler] Outbound handler stop with error: %w", err)
					return
				}
				o.logger.Sugar().Info("[OutboundHandler] Outbound handler stopped")
				return
			case <-ticker.C:
				err := o.handleOutbound(ctx)
				if err != nil {
					o.logger.Sugar().Errorf("[OutboundHandler] Handle Outbound fail: %w", err)
					continue
				}
			}
		}
	}(ctx)
	return nil
}

func (o *OutboundHandler) handleOutbound(ctx context.Context) error {
	pendingOutboundEvents, err := o.loader(ctx)
	if err != nil {
		return err
	}
	eventIds := []uuid.UUID{}
	for _, event := range pendingOutboundEvents {
		err := o.publishOutboundEvent(ctx, event)
		if err != nil {
			o.logger.Sugar().Errorf("[OutboundHandler] Fail to publish pending outbound event to kafka: %w", err)
		} else {
			eventIds = append(eventIds, event.Id)
		}
	}
	err = o.updaterFunc(ctx, eventIds)
	if err != nil {
		return err
	}
	return nil
}

func (o *OutboundHandler) publishOutboundEvent(ctx context.Context, event OutboundEvent) error {
	kafkaMsg := kafka.Event{
		Type:      event.EventType,
		Timestamp: time.Now(),
		Data:      event.Payload,
	}
	err := o.producer.Publish(ctx, event.Topic, event.Id.String(), kafkaMsg)
	if err != nil {
		return err
	}
	return nil
}
