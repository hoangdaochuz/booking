package kafka

import (
	"context"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"go.uber.org/zap"
)

type SagaOrchestratorProducer struct {
	logger   *zap.Logger
	producer *pkgkafka.Producer
}

func NewSagaOrchestratorProducer(brokes []string, logger *zap.Logger) *SagaOrchestratorProducer {
	return &SagaOrchestratorProducer{
		logger:   logger,
		producer: pkgkafka.NewProducer(brokes, []string{"order.events", "booking.events", "notification.events"}, logger),
	}
}

func (s *SagaOrchestratorProducer) PublishEvent(ctx context.Context, topic string, key string, event pkgkafka.Event) error {
	return s.producer.Publish(ctx, topic, key, event)
}
