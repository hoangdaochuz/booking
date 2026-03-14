package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ticketbox/booking/internal/domain"
	"github.com/ticketbox/booking/internal/repository"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	"github.com/ticketbox/pkg/redis"
)

type BookingService struct {
	bookingRepo repository.BookingRepository
	eventClient eventv1.EventServiceClient
	logger      *zap.Logger
	redisClient *redis.RedisClient
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	eventClient eventv1.EventServiceClient,
	logger *zap.Logger,
	redisClient *redis.RedisClient,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		eventClient: eventClient,
		logger:      logger,
		redisClient: redisClient,
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

	// AcquireLock
	var seatIdsReq []string
	acquireKeys := []string{}
	for _, item := range items {
		acquirekey := userID.String() + eventID.String() + item.TicketTierID.String()
		for _, seatId := range item.SeatIDs {
			acquirekey = acquirekey + seatId.String()
			acquireKeys = append(acquireKeys, acquirekey)
			seatIdsReq = append(seatIdsReq, seatId.String())
		}
	}

	var wg sync.WaitGroup
	lockChan := make(chan *redis.RedisLock, len(acquireKeys))
	errChan := make(chan error, len(acquireKeys))
	for _, key := range acquireKeys {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock, err := s.redisClient.AcquireLock(ctx, key, 5*time.Second)
			if err != nil {
				errChan <- fmt.Errorf("fail to acquire lock for seat: %w", err)
				return
			}
			lockChan <- lock
		}()
	}
	wg.Wait()
	close(errChan)
	close(lockChan)

	var lockErr error
	for err := range errChan {
		if err != nil {
			lockErr = err
			break
		}
	}
	if lockErr != nil {
		for lock := range lockChan {
			if lock != nil {
				_ = lock.ReleaseLock(ctx)
			}
		}
		return nil, lockErr
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
		// Why we reduce availability seat here? We doesn't make sure that user complete the booking?
		// var updateErr error
		// switch mode {
		// case "naive":
		// _, updateErr = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
		// 	TierId:        item.TicketTierID.String(),
		// 	QuantityDelta: -item.Quantity,
		// 	Mode:          "naive",
		// })
		// case "pessimistic":
		// 	_, updateErr = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
		// 		TierId:        item.TicketTierID.String(),
		// 		QuantityDelta: -item.Quantity,
		// 		Mode:          "pessimistic",
		// 	})
		// case "optimistic":
		// 	maxRetries := 3
		// 	for attempt := 0; attempt < maxRetries; attempt++ {
		// 		tierInfo, getErr := s.eventClient.GetTicketAvailability(ctx, &eventv1.GetTicketAvailabilityRequest{
		// 			TierId: item.TicketTierID.String(),
		// 		})
		// 		if getErr != nil {
		// 			updateErr = getErr
		// 			break
		// 		}
		// 		_, updateErr = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
		// 			TierId:          item.TicketTierID.String(),
		// 			QuantityDelta:   -item.Quantity,
		// 			ExpectedVersion: tierInfo.Version,
		// 			Mode:            "optimistic",
		// 		})
		// 		if updateErr == nil {
		// 			break
		// 		}
		// 		s.logger.Warn("Optimistic lock retry", zap.Int("attempt", attempt+1))
		// 	}
		// default:
		// 	return nil, fmt.Errorf("unknown booking mode: %s", mode)
		// }

		// if updateErr != nil {
		// 	booking.Status = domain.StatusFailed
		// 	booking.TotalAmountCents = totalCents
		// 	booking.Items = items
		// 	_ = s.bookingRepo.Create(ctx, booking)
		// 	return booking, fmt.Errorf("reserve tickets failed: %w", updateErr)
		// }
	}

	booking.TotalAmountCents = totalCents
	booking.Status = domain.StatusPending
	booking.Items = items

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("save booking: %w", err)
	}

	// Update seat statuses to 'reserved' for all selected seats
	// TODO: Improve later. Should batch update seat status by seatIds
	// for _, item := range items {
	// 	for _, seatID := range item.SeatIDs {
	// 	}
	// }
	_, err := s.eventClient.UpdateBatchSeatStatus(ctx, &eventv1.UpdateBatchSeatStatusRequest{
		SeatIds:   seatIdsReq,
		Status:    "reserved",
		BookingId: booking.ID.String(),
	})
	if err != nil {
		s.logger.Error("Failed to update batch seat status",
			zap.Error(err))
		return nil, fmt.Errorf("update batch seat status failed: %w", err)
	}

	// Release Lock
	for lock := range lockChan {
		if lock != nil {
			err := lock.ReleaseLock(ctx)
			if err != nil {
				s.logger.Error("Fail to release lock", zap.String("key", lock.Key()), zap.Error(err))
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
