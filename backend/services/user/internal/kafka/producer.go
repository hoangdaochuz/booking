package kafka

import (
	"context"
	"encoding/json"
	"time"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/user/internal/domain"
	"go.uber.org/zap"
)

const TopicUserEvents = "user.events"

type UserEventProducer struct {
	producer *pkgkafka.Producer
	logger   *zap.Logger
}

func NewUserEventProducer(brokers []string, logger *zap.Logger) *UserEventProducer {
	producer := pkgkafka.NewProducer(brokers, []string{TopicUserEvents}, logger)
	return &UserEventProducer{producer: producer, logger: logger}
}

func (p *UserEventProducer) PublishUserRegistered(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(map[string]interface{}{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"name":    user.Name,
	})
	if err != nil {
		return err
	}

	event := pkgkafka.Event{
		Type:      "UserRegistered",
		Timestamp: time.Now(),
		Data:      data,
	}

	return p.producer.Publish(ctx, TopicUserEvents, user.ID.String(), event)
}

func (p *UserEventProducer) Close() error {
	return p.producer.Close()
}
