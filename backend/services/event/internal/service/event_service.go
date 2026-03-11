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
	logger    *zap.Logger
}

func NewEventService(eventRepo repository.EventRepository, tierRepo repository.TicketTierRepository, logger *zap.Logger) *EventService {
	return &EventService{eventRepo: eventRepo, tierRepo: tierRepo, logger: logger}
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

	return event, nil
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
