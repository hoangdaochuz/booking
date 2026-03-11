package kafka

import (
	"context"
	"encoding/json"
	"time"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/booking/internal/domain"
	"go.uber.org/zap"
)

const TopicBookingEvents = "booking.events"

type BookingEventProducer struct {
	producer *pkgkafka.Producer
	logger   *zap.Logger
}

func NewBookingEventProducer(brokers []string, logger *zap.Logger) *BookingEventProducer {
	producer := pkgkafka.NewProducer(brokers, []string{TopicBookingEvents}, logger)
	return &BookingEventProducer{producer: producer, logger: logger}
}

func (p *BookingEventProducer) PublishBookingConfirmed(ctx context.Context, booking *domain.Booking, email string) error {
	data, err := json.Marshal(map[string]interface{}{
		"booking_id": booking.ID.String(),
		"user_id":    booking.UserID.String(),
		"event_id":   booking.EventID.String(),
		"email":      email,
		"total":      booking.TotalAmountCents,
	})
	if err != nil {
		return err
	}

	event := pkgkafka.Event{
		Type:      "BookingConfirmed",
		Timestamp: time.Now(),
		Data:      data,
	}
	return p.producer.Publish(ctx, TopicBookingEvents, booking.ID.String(), event)
}

func (p *BookingEventProducer) PublishBookingFailed(ctx context.Context, booking *domain.Booking, email string) error {
	data, err := json.Marshal(map[string]interface{}{
		"booking_id": booking.ID.String(),
		"user_id":    booking.UserID.String(),
		"event_id":   booking.EventID.String(),
		"email":      email,
		"reason":     "insufficient tickets",
	})
	if err != nil {
		return err
	}

	event := pkgkafka.Event{
		Type:      "BookingFailed",
		Timestamp: time.Now(),
		Data:      data,
	}
	return p.producer.Publish(ctx, TopicBookingEvents, booking.ID.String(), event)
}

func (p *BookingEventProducer) PublishBookingCancelled(ctx context.Context, booking *domain.Booking, email string) error {
	data, err := json.Marshal(map[string]interface{}{
		"booking_id": booking.ID.String(),
		"user_id":    booking.UserID.String(),
		"event_id":   booking.EventID.String(),
		"email":      email,
	})
	if err != nil {
		return err
	}

	event := pkgkafka.Event{
		Type:      "BookingCancelled",
		Timestamp: time.Now(),
		Data:      data,
	}
	return p.producer.Publish(ctx, TopicBookingEvents, booking.ID.String(), event)
}

func (p *BookingEventProducer) Close() error {
	return p.producer.Close()
}
