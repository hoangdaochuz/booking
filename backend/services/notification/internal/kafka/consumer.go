package kafka

import (
	"context"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/notification/internal/service"
	"go.uber.org/zap"
)

type NotificationConsumer struct {
	bookingConsumer *pkgkafka.Consumer
	userConsumer    *pkgkafka.Consumer
	service         *service.NotificationService
	logger          *zap.Logger
}

func NewNotificationConsumer(brokers []string, svc *service.NotificationService, logger *zap.Logger) *NotificationConsumer {
	nc := &NotificationConsumer{service: svc, logger: logger}

	nc.bookingConsumer = pkgkafka.NewConsumer(
		brokers, "booking.events", "notification-booking-group",
		nc.handleBookingEvent, logger,
	)
	nc.userConsumer = pkgkafka.NewConsumer(
		brokers, "user.events", "notification-user-group",
		nc.handleUserEvent, logger,
	)

	return nc
}

func (c *NotificationConsumer) Start(ctx context.Context) error {
	errCh := make(chan error, 2)
	go func() { errCh <- c.bookingConsumer.Start(ctx) }()
	go func() { errCh <- c.userConsumer.Start(ctx) }()
	return <-errCh
}

func (c *NotificationConsumer) handleBookingEvent(ctx context.Context, event pkgkafka.Event) error {
	switch event.Type {
	case "BookingConfirmed":
		return c.service.SendBookingConfirmation(ctx, event.Data)
	case "BookingFailed":
		return c.service.SendBookingFailure(ctx, event.Data)
	case "BookingCancelled":
		return c.service.SendBookingCancellation(ctx, event.Data)
	default:
		return nil
	}
}

func (c *NotificationConsumer) handleUserEvent(ctx context.Context, event pkgkafka.Event) error {
	switch event.Type {
	case "UserRegistered":
		return c.service.SendWelcomeEmail(ctx, event.Data)
	default:
		return nil
	}
}
