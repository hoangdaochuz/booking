package service

import (
	"context"

	"github.com/ticketbox/notification/internal/domain"
	"go.uber.org/zap"
)

type NotificationSender interface {
	Send(ctx context.Context, notification *domain.Notification) error
}

type LogSender struct {
	logger *zap.Logger
}

func NewLogSender(logger *zap.Logger) *LogSender {
	return &LogSender{logger: logger}
}

func (s *LogSender) Send(ctx context.Context, n *domain.Notification) error {
	s.logger.Info("NOTIFICATION SENT",
		zap.String("type", n.Type),
		zap.String("recipient", n.Recipient),
		zap.String("channel", n.Channel),
		zap.Any("payload", n.Payload),
	)
	return nil
}
