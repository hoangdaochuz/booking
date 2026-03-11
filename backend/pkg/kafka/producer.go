package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Event struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type Producer struct {
	writers map[string]*kafkago.Writer
	logger  *zap.Logger
}

func NewProducer(brokers []string, topics []string, logger *zap.Logger) *Producer {
	writers := make(map[string]*kafkago.Writer)
	for _, topic := range topics {
		writers[topic] = &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafkago.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafkago.RequireOne,
		}
	}
	return &Producer{writers: writers, logger: logger}
}

func (p *Producer) Publish(ctx context.Context, topic string, key string, event Event) error {
	writer, ok := p.writers[topic]
	if !ok {
		p.logger.Error("Unknown topic", zap.String("topic", topic))
		return fmt.Errorf("unknown topic: %s", topic)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafkago.Message{
		Key:   []byte(key),
		Value: data,
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish", zap.String("topic", topic), zap.Error(err))
		return fmt.Errorf("publish to %s: %w", topic, err)
	}

	p.logger.Debug("Published event", zap.String("topic", topic), zap.String("type", event.Type))
	return nil
}

func (p *Producer) Close() error {
	for _, w := range p.writers {
		if err := w.Close(); err != nil {
			p.logger.Error("Failed to close writer", zap.Error(err))
		}
	}
	return nil
}
