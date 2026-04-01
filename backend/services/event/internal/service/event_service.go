package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ticketbox/event/internal/domain"
	"github.com/ticketbox/event/internal/repository"
)

type EventService struct {
	eventRepo repository.EventRepository
	tierRepo  repository.TicketTierRepository
	seatRepo  repository.SeatRepository
	logger    *zap.Logger
}

func NewEventService(eventRepo repository.EventRepository, tierRepo repository.TicketTierRepository, seatRepo repository.SeatRepository, logger *zap.Logger) *EventService {
	return &EventService{eventRepo: eventRepo, tierRepo: tierRepo, seatRepo: seatRepo, logger: logger}
}

func (s *EventService) CreateEvent(ctx context.Context, title, description, category, venue, location, imageURL string, date time.Time, tiers []domain.TicketTier) (*domain.Event, error) {
	now := time.Now()
	event := &domain.Event{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Category:    category,
		Venue:       venue,
		Location:    location,
		Date:        date,
		ImageURL:    imageURL,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for i := range tiers {
		tiers[i].ID = uuid.New()
		tiers[i].EventID = event.ID
		tiers[i].AvailableQuantity = tiers[i].TotalQuantity
		tiers[i].CreatedAt = now
	}
	event.Tiers = tiers

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}

	// Generate seats for each tier
	if err := s.generateSeatsForEvent(ctx, event, tiers); err != nil {
		s.logger.Error("Failed to generate seats", zap.Error(err))
		// Don't fail the event creation if seat generation fails
		// In production, you might want to rollback or implement retry logic
	}

	return event, nil
}

func (s *EventService) generateSeatsForEvent(ctx context.Context, event *domain.Event, tiers []domain.TicketTier) error {
	var allSeats []*domain.Seat
	now := time.Now()

	for _, tier := range tiers {
		// Generate seats based on tier quantity
		seats := s.generateSeatsForTier(event.ID, tier.ID, tier.Name, tier.TotalQuantity, now)
		allSeats = append(allSeats, seats...)
	}

	if len(allSeats) > 0 {
		if err := s.seatRepo.CreateBatch(ctx, allSeats); err != nil {
			return fmt.Errorf("create seats: %w", err)
		}
		s.logger.Info("Generated seats for event",
			zap.String("event_id", event.ID.String()),
			zap.Int("total_seats", len(allSeats)))
	}

	return nil
}

func (s *EventService) generateSeatsForTier(eventID, tierID uuid.UUID, tierName string, quantity int32, now time.Time) []*domain.Seat {
	seats := make([]*domain.Seat, 0, quantity)

	// Calculate rows and seats per row for layout
	// For simplicity, we'll use a grid layout
	rows := int((quantity + 19) / 20) // Max 20 seats per row
	if rows < 1 {
		rows = 1
	}

	seatNum := 0
	for row := 0; row < rows && seatNum < int(quantity); row++ {
		rowLabel := string(rune('A' + row))
		seatsInRow := int(quantity) - seatNum
		if seatsInRow > 20 {
			seatsInRow = 20
		}

		for i := 0; i < seatsInRow; i++ {
			seats = append(seats, &domain.Seat{
				ID:           uuid.New(),
				EventID:      eventID,
				TicketTierID: tierID,
				Status:       domain.SeatStatusAvailable,
				BookingID:    nil,
				OrderID:      nil,
				Position: domain.Position{
					SectionID: tierName,
					Row:       rowLabel,
					Seat:      i + 1,
					X:         i * 50,
					Y:         row * 50,
				},
				CreatedAt: now,
				UpdatedAt: now,
			})
			seatNum++
		}
	}

	return seats
}

func (s *EventService) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.Event, error) {
	return s.eventRepo.GetByID(ctx, eventID)
}

func (s *EventService) ListEvents(ctx context.Context, category, search string, page, pageSize int) ([]*domain.Event, int, error) {
	return s.eventRepo.List(ctx, category, search, page, pageSize)
}

func (s *EventService) UpdateEvent(ctx context.Context, eventID uuid.UUID, title, description, category, venue, location, imageURL string, date time.Time) (*domain.Event, error) {
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if title != "" {
		event.Title = title
	}
	if description != "" {
		event.Description = description
	}
	if category != "" {
		event.Category = category
	}
	if venue != "" {
		event.Venue = venue
	}
	if location != "" {
		event.Location = location
	}
	if imageURL != "" {
		event.ImageURL = imageURL
	}
	if !date.IsZero() {
		event.Date = date
	}

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}

	return s.eventRepo.GetByID(ctx, eventID)
}

func (s *EventService) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	return s.eventRepo.Delete(ctx, eventID)
}

func (s *EventService) GetTicketAvailability(ctx context.Context, tierID uuid.UUID) (*domain.TicketTier, error) {
	return s.tierRepo.GetByID(ctx, tierID)
}

func (s *EventService) UpdateTicketAvailability(ctx context.Context, tierID uuid.UUID, delta int32, expectedVersion int32, mode string) (*domain.TicketTier, error) {
	switch mode {
	case "naive":
		if err := s.tierRepo.UpdateAvailabilityNoLock(ctx, tierID, delta); err != nil {
			return nil, err
		}
		return s.tierRepo.GetByID(ctx, tierID)
	case "pessimistic":
		return s.tierRepo.UpdateAvailabilityPessimistic(ctx, tierID, delta)
	case "optimistic":
		return s.tierRepo.UpdateAvailability(ctx, tierID, delta, expectedVersion)
	default:
		return s.tierRepo.UpdateAvailabilityPessimistic(ctx, tierID, delta)
	}
}

type UpdateTicketAvailabilityRequest struct {
	tierId        uuid.UUID
	deltaQuantity int
}
type UpdateBatchTicketAvailabilityRequest struct {
	Data []UpdateBatchTicketAvailabilityRequest
}

func (s *EventService) UpdateBatchTicketAvailability(ctx context.Context, req *UpdateBatchTicketAvailabilityRequest) error {
	return nil
}

func (s *EventService) GetSeats(ctx context.Context, eventID uuid.UUID, tierID *uuid.UUID) ([]*domain.Seat, error) {
	return s.seatRepo.GetByEventID(ctx, eventID, tierID)
}

func (s *EventService) UpdateSeatStatus(ctx context.Context, seatID uuid.UUID, status domain.SeatStatus, bookingID *uuid.UUID) (*domain.Seat, error) {
	return s.seatRepo.UpdateStatus(ctx, seatID, status, bookingID)
}

func (s *EventService) UpdateBatchSeatStatus(ctx context.Context, seatIds []uuid.UUID, status domain.SeatStatus, bookingID *uuid.UUID) error {
	return s.seatRepo.UpdateStatusBatch(ctx, seatIds, status, bookingID)
}
