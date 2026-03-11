package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ticketbox/event/internal/domain"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("version conflict")
	ErrInsufficientTickets = errors.New("insufficient tickets")
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	List(ctx context.Context, category, search string, page, pageSize int) ([]*domain.Event, int, error)
	Update(ctx context.Context, event *domain.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TicketTierRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketTier, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]domain.TicketTier, error)
	UpdateAvailability(ctx context.Context, tierID uuid.UUID, delta int32, expectedVersion int32) (*domain.TicketTier, error)
	UpdateAvailabilityNoLock(ctx context.Context, tierID uuid.UUID, delta int32) error
	UpdateAvailabilityPessimistic(ctx context.Context, tierID uuid.UUID, delta int32) (*domain.TicketTier, error)
}
