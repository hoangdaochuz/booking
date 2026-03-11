package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ticketbox/booking/internal/domain"
)

var ErrNotFound = errors.New("not found")

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Booking, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
}
