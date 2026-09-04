package kafka

import (
	"context"
	"encoding/json"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/pkg/topics"
	"github.com/ticketbox/scheduler/internal/domain"
	service_scheduler "github.com/ticketbox/scheduler/internal/service/scheduler"
	"go.uber.org/zap"
)

type SchedulerConfigConsumer struct {
	service                      *service_scheduler.SchedulerService
	logger                       *zap.Logger
	schedulerConfigEventConsumer *pkgkafka.Consumer
}

func NewSchedulerConfigConsumer(broker []string, service *service_scheduler.SchedulerService, logger *zap.Logger) *SchedulerConfigConsumer {
	result := &SchedulerConfigConsumer{
		service: service,
		logger:  logger,
	}
	result.schedulerConfigEventConsumer = pkgkafka.NewConsumer(
		broker,
		topics.TopicSchedulerConfigEvent,
		"scheduler-config-event-group",
		result.ConsumerHandler,
		logger,
	)
	return result
}

func (o *SchedulerConfigConsumer) ConsumerHandler(ctx context.Context, event pkgkafka.Event) error {
	switch event.Type {
	case topics.TypeSchedulerConfigChanged:
		var payload domain.SchedulerConfig
		err := json.Unmarshal(event.Data, &payload)
		if err != nil {
			o.logger.Sugar().Error("[SchedulerConfigConsumer][SchedulerConfigChanged] Fail to unmarshal payload", zap.Error(err))
		}
		return o.service.HandleSchedulerConfigChanged(ctx, payload)
	default:
		o.logger.Sugar().Warn("[SchedulerConfigConsumer] The type of event is out of available event types")
	}
	return nil
}

func (o *SchedulerConfigConsumer) Start(ctx context.Context) error {
	return o.schedulerConfigEventConsumer.Start(ctx)
}
