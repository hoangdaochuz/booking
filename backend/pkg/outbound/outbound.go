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
	o.logger.Sugar().Infoln("[OutboundHandler] Outbound handler is starting")
	go func(ctx context.Context) {
		// Stop must live inside the goroutine: deferring it in Start()
		// stops the ticker as soon as Start returns, so ticker.C never
		// fires and the loop below blocks forever.
		defer ticker.Stop()
		o.logger.Sugar().Infoln("[OutboundHandler] Outbound handler is running")
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
				o.logger.Sugar().Info("[OutboundHandler] handling outbound events")
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
	o.logger.Sugar().Infof("[OutboundHandler] Found %d pending outbound events", len(pendingOutboundEvents))
	eventIds := []uuid.UUID{}
	if len(pendingOutboundEvents) == 0 {
		o.logger.Sugar().Info("[OutboundHandler] No pending outbound events to process")
		return nil
	}
	for _, event := range pendingOutboundEvents {
		err := o.publishOutboundEvent(ctx, event)
		if err != nil {
			o.logger.Sugar().Errorf("[OutboundHandler] Fail to publish pending outbound event to kafka: %w", err)
		} else {
			eventIds = append(eventIds, event.Id)
		}
	}
	o.logger.Sugar().Infof("[OutboundHandler] Publishing %d outbound events", len(eventIds))
	if len(eventIds) == 0 {
		o.logger.Sugar().Info("[OutboundHandler] No outbound events to mark as published")
		return nil
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
