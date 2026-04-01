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
	ErrSeatNotAvailable    = errors.New("seat not available")
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	List(ctx context.Context, category, search string, page, pageSize int) ([]*domain.Event, int, error)
	Update(ctx context.Context, event *domain.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UpdateTicketAvailabilityRequest struct {
	tierId        uuid.UUID
	deltaQuantity int
}
type UpdateBatchTicketAvailabilityRequest struct {
	Data []UpdateBatchTicketAvailabilityRequest
}

type TicketTierRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketTier, error)
	GetByEventID(ctx context.Context, eventID uuid.UUID) ([]domain.TicketTier, error)
	UpdateAvailability(ctx context.Context, tierID uuid.UUID, delta int32, expectedVersion int32) (*domain.TicketTier, error)
	UpdateAvailabilityNoLock(ctx context.Context, tierID uuid.UUID, delta int32) error
	UpdateAvailabilityPessimistic(ctx context.Context, tierID uuid.UUID, delta int32) (*domain.TicketTier, error)
	UpdateBatchTicketAvailability(ctx context.Context, req *UpdateBatchTicketAvailabilityRequest) error
}

type SeatRepository interface {
	// Create inserts a new seat into the database
	Create(ctx context.Context, seat *domain.Seat) error

	// CreateBatch inserts multiple seats in a single transaction
	CreateBatch(ctx context.Context, seats []*domain.Seat) error

	// GetByID retrieves a seat by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Seat, error)

	// GetByEventID retrieves all seats for an event, optionally filtered by tier ID
	GetByEventID(ctx context.Context, eventID uuid.UUID, tierID *uuid.UUID) ([]*domain.Seat, error)

	// GetByTierID retrieves all seats for a ticket tier
	GetByTierID(ctx context.Context, tierID uuid.UUID) ([]*domain.Seat, error)

	// UpdateStatus updates the status of a seat, optionally associating it with a booking
	UpdateStatus(ctx context.Context, seatID uuid.UUID, status domain.SeatStatus, bookingID *uuid.UUID) (*domain.Seat, error)

	// UpdateStatusBatch updates multiple seats' statuses in a single transaction
	UpdateStatusBatch(ctx context.Context, seatIDs []uuid.UUID, status domain.SeatStatus, bookingID *uuid.UUID) error

	// GetAvailableSeats retrieves all available seats for an event
	GetAvailableSeats(ctx context.Context, eventID uuid.UUID) ([]*domain.Seat, error)

	// GetAvailableSeatsByTier retrieves available seats for a specific ticket tier
	GetAvailableSeatsByTier(ctx context.Context, tierID uuid.UUID) ([]*domain.Seat, error)
}
