package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ticketbox/booking/internal/domain"
	"github.com/ticketbox/booking/internal/repository"
	"github.com/ticketbox/booking/internal/service"
	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
)

type BookingServer struct {
	bookingv1.UnimplementedBookingServiceServer
	service *service.BookingService
	logger  *zap.Logger
}

func NewBookingServer(svc *service.BookingService, logger *zap.Logger) *BookingServer {
	return &BookingServer{service: svc, logger: logger}
}

func (s *BookingServer) CreateBooking(ctx context.Context, req *bookingv1.CreateBookingRequest) (*bookingv1.BookingDetail, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event ID")
	}

	var items []domain.BookingItem
	for _, item := range req.Items {
		tierID, err := uuid.Parse(item.TicketTierId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid tier ID")
		}

		var seatIDs []uuid.UUID
		for _, seatID := range item.SeatIds {
			id, err := uuid.Parse(seatID)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid seat ID")
			}
			seatIDs = append(seatIDs, id)
		}

		items = append(items, domain.BookingItem{
			TicketTierID: tierID,
			Quantity:     item.Quantity,
			SeatIDs:      seatIDs,
		})
	}

	mode := req.BookingMode
	if mode == "" {
		mode = "pessimistic"
	}

	booking, err := s.service.CreateBooking(ctx, userID, eventID, items, mode)
	if err != nil {
		if booking != nil && booking.Status == domain.StatusFailed {
			return toBookingDetail(booking), nil
		}
		s.logger.Error("CreateBooking failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "booking failed")
	}

	return toBookingDetail(booking), nil
}

func (s *BookingServer) GetBooking(ctx context.Context, req *bookingv1.GetBookingRequest) (*bookingv1.BookingDetail, error) {
	bookingID, err := uuid.Parse(req.BookingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid booking ID")
	}

	booking, err := s.service.GetBooking(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "booking not found")
		}
		return nil, status.Error(codes.Internal, "failed to get booking")
	}

	return toBookingDetail(booking), nil
}

func (s *BookingServer) ListUserBookings(ctx context.Context, req *bookingv1.ListUserBookingsRequest) (*bookingv1.ListUserBookingsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	bookings, total, err := s.service.ListUserBookings(ctx, userID, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list bookings")
	}

	var details []*bookingv1.BookingDetail
	for _, b := range bookings {
		details = append(details, toBookingDetail(b))
	}

	return &bookingv1.ListUserBookingsResponse{
		Bookings: details,
		Total:    int32(total),
	}, nil
}

func (s *BookingServer) CancelBooking(ctx context.Context, req *bookingv1.CancelBookingRequest) (*bookingv1.BookingDetail, error) {
	bookingID, err := uuid.Parse(req.BookingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid booking ID")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	booking, err := s.service.CancelBooking(ctx, bookingID, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cancel failed")
	}

	return toBookingDetail(booking), nil
}

func toBookingDetail(b *domain.Booking) *bookingv1.BookingDetail {
	detail := &bookingv1.BookingDetail{
		Id:               b.ID.String(),
		UserId:           b.UserID.String(),
		EventId:          b.EventID.String(),
		Status:           string(b.Status),
		TotalAmountCents: b.TotalAmountCents,
		CreatedAt:        timestamppb.New(b.CreatedAt),
	}
	for _, item := range b.Items {
		var seatIDs []string
		for _, id := range item.SeatIDs {
			seatIDs = append(seatIDs, id.String())
		}
		detail.Items = append(detail.Items, &bookingv1.BookingItem{
			Id:             item.ID.String(),
			TicketTierId:   item.TicketTierID.String(),
			TierName:       item.TierName,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			SeatIds:        seatIDs,
		})
	}
	return detail
}
