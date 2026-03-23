package kafka

import (
	"context"
	"encoding/json"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, event Event) error

type Consumer struct {
	reader  *kafkago.Reader
	logger  *zap.Logger
	handler MessageHandler
}

func NewConsumer(brokers []string, topic string, groupID string, handler MessageHandler, logger *zap.Logger) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{reader: reader, logger: logger, handler: handler}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Consumer started", zap.String("topic", c.reader.Config().Topic))

	for {
		select {
		case <-ctx.Done():
			return c.reader.Close()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return err
				}
				c.logger.Error("Read message failed", zap.Error(err))
				continue
			}

			var event Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Error("Unmarshal event failed", zap.Error(err))
				continue
			}

			if err := c.handler(ctx, event); err != nil {
				c.logger.Error("Handle event failed",
					zap.String("type", event.Type), zap.Error(err))
			}
		}
	}
}
