package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ticketbox/event/internal/domain"
	"github.com/ticketbox/event/internal/repository"
	"github.com/ticketbox/event/internal/service"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
)

type EventServer struct {
	eventv1.UnimplementedEventServiceServer
	service *service.EventService
	logger  *zap.Logger
}

func NewEventServer(svc *service.EventService, logger *zap.Logger) *EventServer {
	return &EventServer{service: svc, logger: logger}
}

func (s *EventServer) CreateEvent(ctx context.Context, req *eventv1.CreateEventRequest) (*eventv1.EventDetail, error) {
	var tiers []domain.TicketTier
	for _, t := range req.Tiers {
		tiers = append(tiers, domain.TicketTier{
			Name:          t.Name,
			PriceCents:    t.PriceCents,
			TotalQuantity: t.TotalQuantity,
		})
	}

	date := req.Date.AsTime()
	event, err := s.service.CreateEvent(ctx, req.Title, req.Description, req.Category, req.Venue, req.Location, req.ImageUrl, date, tiers)
	if err != nil {
		s.logger.Error("CreateEvent failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create event")
	}

	return toEventDetail(event), nil
}

func (s *EventServer) GetEvent(ctx context.Context, req *eventv1.GetEventRequest) (*eventv1.EventDetail, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event ID")
	}

	event, err := s.service.GetEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "event not found")
		}
		return nil, status.Error(codes.Internal, "failed to get event")
	}

	return toEventDetail(event), nil
}

func (s *EventServer) ListEvents(ctx context.Context, req *eventv1.ListEventsRequest) (*eventv1.ListEventsResponse, error) {
	events, total, err := s.service.ListEvents(ctx, req.Category, req.Search, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	var details []*eventv1.EventDetail
	for _, e := range events {
		details = append(details, toEventDetail(e))
	}

	return &eventv1.ListEventsResponse{
		Events:   details,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *EventServer) UpdateEvent(ctx context.Context, req *eventv1.UpdateEventRequest) (*eventv1.EventDetail, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event ID")
	}

	date := req.Date.AsTime()
	event, err := s.service.UpdateEvent(ctx, eventID, req.Title, req.Description, req.Category, req.Venue, req.Location, req.ImageUrl, date)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update event")
	}

	return toEventDetail(event), nil
}

func (s *EventServer) DeleteEvent(ctx context.Context, req *eventv1.DeleteEventRequest) (*eventv1.DeleteEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event ID")
	}

	if err := s.service.DeleteEvent(ctx, eventID); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete event")
	}

	return &eventv1.DeleteEventResponse{}, nil
}

func (s *EventServer) GetTicketAvailability(ctx context.Context, req *eventv1.GetTicketAvailabilityRequest) (*eventv1.TicketAvailabilityResponse, error) {
	tierID, err := uuid.Parse(req.TierId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tier ID")
	}

	tier, err := s.service.GetTicketAvailability(ctx, tierID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "tier not found")
	}

	return &eventv1.TicketAvailabilityResponse{
		TierId:            tier.ID.String(),
		AvailableQuantity: tier.AvailableQuantity,
		Version:           tier.Version,
		PriceCents:        tier.PriceCents,
	}, nil
}

func (s *EventServer) UpdateTicketAvailability(ctx context.Context, req *eventv1.UpdateTicketAvailabilityRequest) (*eventv1.TicketTier, error) {
	tierID, err := uuid.Parse(req.TierId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tier ID")
	}

	tier, err := s.service.UpdateTicketAvailability(ctx, tierID, req.QuantityDelta, req.ExpectedVersion, req.Mode)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, status.Error(codes.Aborted, "version conflict - retry")
		}
		if errors.Is(err, repository.ErrInsufficientTickets) {
			return nil, status.Error(codes.FailedPrecondition, "insufficient tickets")
		}
		return nil, status.Error(codes.Internal, "failed to update availability")
	}

	return toTicketTier(tier), nil
}

func toEventDetail(e *domain.Event) *eventv1.EventDetail {
	detail := &eventv1.EventDetail{
		Id:          e.ID.String(),
		Title:       e.Title,
		Description: e.Description,
		Category:    e.Category,
		Venue:       e.Venue,
		Location:    e.Location,
		Date:        timestamppb.New(e.Date),
		ImageUrl:    e.ImageURL,
		Status:      e.Status,
		CreatedAt:   timestamppb.New(e.CreatedAt),
	}
	for _, t := range e.Tiers {
		detail.Tiers = append(detail.Tiers, toTicketTier(&t))
	}
	return detail
}

func toTicketTier(t *domain.TicketTier) *eventv1.TicketTier {
	return &eventv1.TicketTier{
		Id:                t.ID.String(),
		EventId:           t.EventID.String(),
		Name:              t.Name,
		PriceCents:        t.PriceCents,
		TotalQuantity:     t.TotalQuantity,
		AvailableQuantity: t.AvailableQuantity,
		Version:           t.Version,
	}
}
