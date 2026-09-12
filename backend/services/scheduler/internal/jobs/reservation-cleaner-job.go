package cronjob

import (
	"context"
	"errors"
	"fmt"
	"time"

	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	pkgredis "github.com/ticketbox/pkg/redis"
	"go.uber.org/zap"
)

type ReservationCleanerJob struct {
	eventClient   eventv1.EventServiceClient
	bookingClient bookingv1.BookingServiceClient
	redisLock     *pkgredis.RedisClient
	logger        *zap.Logger
	paymentClient paymentv1.PaymentServiceClient
}

func NewReservationCleanerJob(
	eventClient eventv1.EventServiceClient,
	bookingClient bookingv1.BookingServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	redisLock *pkgredis.RedisClient,
	logger *zap.Logger,
) *ReservationCleanerJob {
	return &ReservationCleanerJob{
		eventClient:   eventClient,
		bookingClient: bookingClient,
		paymentClient: paymentClient,
		redisLock:     redisLock,
		logger:        logger,
	}
}

func (r *ReservationCleanerJob) Name() string {
	return "reservation-cleaner-job"
}

const (
	lockTTL = 30

	// Status/action values cross-referenced with the other services' domains:
	// payment domain PaymentSuccess, booking domain StatusConfirmed/StatusExpired,
	// seat domain SeatStatusBooked, ReserveSeatAction COMPENSATE_SEAT.
	paymentStatusSuccess   = "success"
	paymentStatusPending   = "pending"
	bookingStatusConfirmed = "CONFIRMED"
	bookingStatusExpired   = "EXPIRED"
	seatStatusBooked       = "booked"
	compensateSeatAction   = "compensate"
)

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
		if lock == nil {
			return
		}
		err := lock.ReleaseLock(ctx)
		if err != nil {
			r.logger.Error("[reservation-cleaner-job] Fail to release lock for job", zap.String("job_name", r.Name()), zap.Error(err))
		}
	}()

	reservedExpiredSeatsRes, err := r.eventClient.GetReservedExpiredSeats(ctx, &eventv1.GetReservedExpiredSeatsReq{})
	if err != nil {
		r.logger.Error("[reservation-cleaner-job] Fail to get reserved expired seats", zap.Error(err))
		return err
	}

	if len(reservedExpiredSeatsRes.BookingSeatIdsMap) == 0 {
		r.logger.Info("[reservation-cleaner-job] There is not any stale reserved seats, return job")
		return nil
	}

	bookingIds := make([]string, 0, len(reservedExpiredSeatsRes.BookingSeatIdsMap))
	for bookingId := range reservedExpiredSeatsRes.BookingSeatIdsMap {
		bookingIds = append(bookingIds, bookingId)
	}

	payments, err := r.paymentClient.GetPaymentsByBookingIds(ctx, &paymentv1.GetPaymentsByBookingIdsReq{
		BookingIds: bookingIds,
	})
	if err != nil {
		r.logger.Error("[reservation-cleaner-job] fail to get payments by booking ids", zap.Error(err))
		return err
	}

	// A booking counts as paid when at least one of its payments succeeded.
	// Bookings with no payment record at all (saga paused before payment
	// creation, or payment creation failed) fall through to the expire path.
	paidBookings := make(map[string]bool, len(payments.Payments))
	bookingIdPaymentIdMap := make(map[string]string, len(payments.Payments))
	pendingPayments := make(map[string]bool, len(payments.Payments))
	for _, payment := range payments.Payments {
		if payment.Status == paymentStatusSuccess {
			paidBookings[payment.BookingId] = true
		}
		if payment.Status == paymentStatusPending {
			pendingPayments[payment.Id] = true
		}
		bookingIdPaymentIdMap[payment.BookingId] = payment.Id
	}

	// Reconcile each booking independently: one failing booking must not
	// block the cleanup of the others, so errors are collected and reported
	// after the loop.
	var errs []error
	for bookingId, seatIds := range reservedExpiredSeatsRes.BookingSeatIdsMap {
		if err := r.reconcileBooking(ctx, bookingId, seatIds.SeatIds, paidBookings[bookingId], bookingIdPaymentIdMap[bookingId], pendingPayments); err != nil {
			r.logger.Error("[reservation-cleaner-job] Fail to reconcile booking",
				zap.String("booking_id", bookingId), zap.Error(err))
			errs = append(errs, fmt.Errorf("booking %s: %w", bookingId, err))
		}
	}

	return errors.Join(errs...)
}

func (r *ReservationCleanerJob) reconcileBooking(ctx context.Context, bookingId string, seatIds []string, paid bool, paymentId string, pendingPaymentsMap map[string]bool) error {
	if paid {
		// Payment succeeded but the saga never resumed past the payment
		// webhook: finish its remaining steps by hand.
		_, err := r.eventClient.UpdateBatchSeatStatus(ctx, &eventv1.UpdateBatchSeatStatusRequest{
			SeatIds:   seatIds,
			Status:    seatStatusBooked,
			BookingId: bookingId,
		})
		if err != nil {
			return fmt.Errorf("mark seats booked: %w", err)
		}

		_, err = r.bookingClient.UpdateBookingStatusById(ctx, &bookingv1.UpdateBookingStatusByIdReq{
			Id:     bookingId,
			Status: bookingStatusConfirmed,
		})
		if err != nil {
			return fmt.Errorf("confirm booking: %w", err)
		}
		return nil
	}

	// No successful payment: release the seats back to available (clears
	// reserved_by_booking_id and reservation_expired_at) and expire the booking.
	_, err := r.eventClient.ReservedOrCompensateBatchSeats(ctx, &eventv1.ReservedOrCompensateBatchSeatsReq{
		SeatIds:             seatIds,
		Action:              compensateSeatAction,
		ReservedByBookingId: bookingId,
	})
	if err != nil {
		return fmt.Errorf("release seats: %w", err)
	}

	_, err = r.bookingClient.UpdateBookingStatusById(ctx, &bookingv1.UpdateBookingStatusByIdReq{
		Id:     bookingId,
		Status: bookingStatusExpired,
	})
	if err != nil {
		return fmt.Errorf("expire booking: %w", err)
	}

	if pendingPaymentsMap[paymentId] {
		_, err = r.paymentClient.UpdatePaymentStatus(ctx, &paymentv1.UpdatePaymentStatusRequest{Id: paymentId, Status: "timeout"})
		if err != nil {
			return fmt.Errorf("fail to update pending payment status: %w", err)
		}
	}
	return nil
}
