package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ticketbox/event/internal/domain"
	pkgkafka "github.com/ticketbox/pkg/kafka"
	"go.uber.org/zap"
)

const TopicEventEvents = "event.events"

type EventEventProducer struct {
	producer *pkgkafka.Producer
	logger   *zap.Logger
}

func NewEventEventProducer(brokers []string, logger *zap.Logger) *EventEventProducer {
	producer := pkgkafka.NewProducer(brokers, []string{TopicEventEvents}, logger)
	return &EventEventProducer{producer: producer, logger: logger}
}

func (p *EventEventProducer) PublishEventCreated(ctx context.Context, event *domain.Event) error {
	data, err := json.Marshal(map[string]interface{}{
		"event_id": event.ID.String(),
		"title":    event.Title,
		"category": event.Category,
	})
	if err != nil {
		return err
	}

	evt := pkgkafka.Event{
		Type:      "EventCreated",
		Timestamp: time.Now(),
		Data:      data,
	}

	return p.producer.Publish(ctx, TopicEventEvents, event.ID.String(), evt)
}

func (p *EventEventProducer) Close() error {
	return p.producer.Close()
}
