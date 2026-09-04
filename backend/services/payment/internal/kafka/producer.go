package kafka

import (
	"context"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/pkg/topics"
	"go.uber.org/zap"
)

type PaymentProducer struct {
	logger   *zap.Logger
	producer *pkgkafka.Producer
}

func NewPaymentProducer(brokers []string, logger *zap.Logger) *PaymentProducer {
	return &PaymentProducer{
		logger:   logger,
		producer: pkgkafka.NewProducer(brokers, []string{topics.TopicPaymentEvents}, logger),
	}
}

func (p *PaymentProducer) PublishEvent(ctx context.Context, topic string, key string, event pkgkafka.Event) error {
	return p.producer.Publish(ctx, topic, key, event)
}
