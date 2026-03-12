package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	"github.com/ticketbox/booking/internal/domain"
	"github.com/ticketbox/booking/internal/repository"
)

type BookingService struct {
	bookingRepo repository.BookingRepository
	eventClient eventv1.EventServiceClient
	logger      *zap.Logger
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	eventClient eventv1.EventServiceClient,
	logger *zap.Logger,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		eventClient: eventClient,
		logger:      logger,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, userID, eventID uuid.UUID, items []domain.BookingItem, mode string) (*domain.Booking, error) {
	now := time.Now()
	booking := &domain.Booking{
		ID:        uuid.New(),
		UserID:    userID,
		EventID:   eventID,
		Status:    domain.StatusPending,
		Version:   1,
		Items:     items,
		CreatedAt: now,
	}

	var totalCents int64
	for i, item := range items {
		items[i].ID = uuid.New()
		items[i].BookingID = booking.ID

		// Validate that seat IDs count matches quantity
		if len(item.SeatIDs) != int(item.Quantity) {
			return nil, fmt.Errorf("seat IDs count (%d) must match quantity (%d)", len(item.SeatIDs), item.Quantity)
		}

		tier, err := s.eventClient.GetTicketAvailability(ctx, &eventv1.GetTicketAvailabilityRequest{
			TierId: item.TicketTierID.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("get tier availability: %w", err)
		}

		items[i].UnitPriceCents = tier.PriceCents
		totalCents += tier.PriceCents * int64(item.Quantity)

		var updateErr error
		switch mode {
		case "naive":
			_, updateErr = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
				TierId:        item.TicketTierID.String(),
				QuantityDelta: -item.Quantity,
				Mode:          "naive",
			})
		case "pessimistic":
			_, updateErr = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
				TierId:        item.TicketTierID.String(),
				QuantityDelta: -item.Quantity,
				Mode:          "pessimistic",
			})
		case "optimistic":
			maxRetries := 3
			for attempt := 0; attempt < maxRetries; attempt++ {
				tierInfo, getErr := s.eventClient.GetTicketAvailability(ctx, &eventv1.GetTicketAvailabilityRequest{
					TierId: item.TicketTierID.String(),
				})
				if getErr != nil {
					updateErr = getErr
					break
				}
				_, updateErr = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
					TierId:          item.TicketTierID.String(),
					QuantityDelta:   -item.Quantity,
					ExpectedVersion: tierInfo.Version,
					Mode:            "optimistic",
				})
				if updateErr == nil {
					break
				}
				s.logger.Warn("Optimistic lock retry", zap.Int("attempt", attempt+1))
			}
		default:
			return nil, fmt.Errorf("unknown booking mode: %s", mode)
		}

		if updateErr != nil {
			booking.Status = domain.StatusFailed
			booking.TotalAmountCents = totalCents
			booking.Items = items
			_ = s.bookingRepo.Create(ctx, booking)
			return booking, fmt.Errorf("reserve tickets failed: %w", updateErr)
		}
	}

	booking.TotalAmountCents = totalCents
	booking.Status = domain.StatusConfirmed
	booking.Items = items

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("save booking: %w", err)
	}

	// Update seat statuses to 'booked' for all selected seats
	for _, item := range items {
		for _, seatID := range item.SeatIDs {
			_, err := s.eventClient.UpdateSeatStatus(ctx, &eventv1.UpdateSeatStatusRequest{
				SeatId:    seatID.String(),
				Status:    "booked",
				BookingId: booking.ID.String(),
			})
			if err != nil {
				s.logger.Error("Failed to update seat status",
					zap.String("seat_id", seatID.String()),
					zap.Error(err))
				// Don't fail the booking if seat update fails - log and continue
			}
		}
	}

	return booking, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	for _, item := range booking.Items {
		_, err := s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
			TierId:        item.TicketTierID.String(),
			QuantityDelta: item.Quantity,
			Mode:          "pessimistic",
		})
		if err != nil {
			s.logger.Error("Failed to restore tickets", zap.Error(err))
		}
	}

	booking.Status = domain.StatusCancelled
	if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.StatusCancelled); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) GetBooking(ctx context.Context, bookingID uuid.UUID) (*domain.Booking, error) {
	return s.bookingRepo.GetByID(ctx, bookingID)
}

func (s *BookingService) ListUserBookings(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Booking, int, error) {
	return s.bookingRepo.ListByUserID(ctx, userID, page, pageSize)
}
