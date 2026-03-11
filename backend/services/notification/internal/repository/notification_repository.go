package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ticketbox/notification/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	MarkSent(ctx context.Context, id uuid.UUID) error
}
