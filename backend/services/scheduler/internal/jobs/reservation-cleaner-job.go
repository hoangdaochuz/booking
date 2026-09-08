package cronjob

import (
	"context"
	"errors"
	"time"

	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	pkgredis "github.com/ticketbox/pkg/redis"
	"go.uber.org/zap"
)

type ReservationCleanerJob struct {
	eventClient   eventv1.EventServiceClient
	bookingClient bookingv1.BookingServiceClient
	redisLock     *pkgredis.RedisClient
	logger        *zap.Logger
}

func NewReservationCleanerJob(
	eventClient eventv1.EventServiceClient,
	bookingClient bookingv1.BookingServiceClient,
	redisLock *pkgredis.RedisClient,
	logger *zap.Logger,
) *ReservationCleanerJob {
	return &ReservationCleanerJob{
		eventClient:   eventClient,
		bookingClient: bookingClient,
		redisLock:     redisLock,
		logger:        logger,
	}
}

func (r *ReservationCleanerJob) Name() string {
	return "reservation-cleaner-job"
}

const lockTTL = 30

func (r *ReservationCleanerJob) Run(ctx context.Context) error {
	// Acquire lock
	lockKey := r.Name()
	lock, err := r.redisLock.AcquireLock(ctx, lockKey, lockTTL*time.Second)
	if err != nil {
		if errors.Is(err, pkgredis.ErrLockAlreadyAcquire) {
			r.logger.Info("[reservation-cleaner-job] Job already acquired lock by other process, returning", zap.String("job_name", r.Name()))
			return nil
		}
		// error when try acquire lock -> still continue
		r.logger.Warn("[reservation-cleaner-job] Fail to acquire lock, still continue job", zap.Error(err))
	}

	defer func() {
		err := lock.ReleaseLock(ctx)
		if err != nil {
			r.logger.Error("[reservation-cleaner-job] Fail to release lock for job", zap.String("job_name", r.Name()), zap.Error(err))
		}
	}()

	undoReservedExpiredSeatsRes, err := r.eventClient.UndoReservedExpiredSeats(ctx, &eventv1.UndoReservedExpiredSeatsReq{})
	if err != nil {
		r.logger.Error("[reservation-cleaner-job] Fail to undo reserved expired seats", zap.Error(err))
		return err
	}

	if len(undoReservedExpiredSeatsRes.BookingSeatIdsMap) == 0 {
		r.logger.Info("[reservation-cleaner-job] There is not any stale reserved seats, return job")
		return nil
	}

	bookingIds := make([]string, 0, len(undoReservedExpiredSeatsRes.BookingSeatIdsMap))
	for bookingId := range undoReservedExpiredSeatsRes.BookingSeatIdsMap {
		bookingIds = append(bookingIds, bookingId)
	}

	_, err = r.bookingClient.UpdateBookingStatusByIds(ctx, &bookingv1.UpdateBookingStatusByIdsReq{BookingIds: bookingIds, Status: "EXPIRED"})
	return err
}
