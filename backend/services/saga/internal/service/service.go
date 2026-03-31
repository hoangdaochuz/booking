package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	sagapkg "github.com/ticketbox/pkg/saga"
	pkgtyped "github.com/ticketbox/pkg/typed"
	"github.com/ticketbox/saga/internal/domain"
	"github.com/ticketbox/saga/internal/registry"
	"github.com/ticketbox/saga/internal/repository"
	"go.uber.org/zap"
)

type SagaService struct {
	logger           *zap.Logger
	repo             repository.SagaRepositoryInterface
	sagaStepRegistry *registry.SagaStepRegistry
	bookingClient    bookingv1.BookingServiceClient
	paymentClient    paymentv1.PaymentServiceClient
	eventClient      eventv1.EventServiceClient
}

func NewSagaService(logger *zap.Logger,
	repo repository.SagaRepositoryInterface,
	sagaStepRegistry *registry.SagaStepRegistry,
	bookingClient bookingv1.BookingServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	eventClient eventv1.EventServiceClient) *SagaService {
	return &SagaService{
		logger:           logger,
		repo:             repo,
		sagaStepRegistry: sagaStepRegistry,
		bookingClient:    bookingClient,
		paymentClient:    paymentClient,
		eventClient:      eventClient,
	}
}

type CreateSagaRequest struct {
	Name      string
	BookingId uuid.UUID
	Steps     []domain.SagaStep
}

type UpdateSagaRequest struct {
	Id               uuid.UUID
	Name             string
	Status           domain.SagaStatus
	CurrentStepIndex int
}

func (s *SagaService) CreateSaga(ctx context.Context, req *CreateSagaRequest) (*domain.Saga, error) {
	sagaCreate := domain.Saga{
		ID:               uuid.New(),
		BookingID:        req.BookingId,
		Name:             req.Name,
		Steps:            req.Steps,
		Status:           domain.SAGA_PENDING,
		CurrentStepIndex: 0,
	}
	err := s.repo.Create(ctx, &sagaCreate)
	if err != nil {
		return nil, err
	}
	return &sagaCreate, nil
}

func (s *SagaService) UpdateSaga(ctx context.Context, req *UpdateSagaRequest) error {
	sagaUpdate := domain.Saga{
		ID:               req.Id,
		Name:             req.Name,
		Status:           req.Status,
		CurrentStepIndex: req.CurrentStepIndex,
	}
	return s.repo.UpdateSaga(ctx, &sagaUpdate, nil)
}

func (s *SagaService) GetSagaById(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {

	return s.repo.GetSagaById(ctx, id)
}

func (s *SagaService) GetSagaByBookingId(ctx context.Context, id uuid.UUID) (*domain.Saga, error) {
	return s.repo.GetSagaByBookingId(ctx, id)
}

// type RegisterSagaStepsRequest struct{}

type StartOrderSagaRequest struct {
	BookingId  string
	SeatIds    []string
	UserId     string
	TotalCents int
}

func (s *SagaService) InitializeSagaHandler(ctx context.Context, req *StartOrderSagaRequest) (*registry.SagaHandler, error) {
	sagaId := uuid.New()
	bookingUUID, _ := uuid.Parse(req.BookingId)
	orderSaga := &domain.Saga{
		ID:               sagaId,
		BookingID:        bookingUUID,
		Name:             "ORDER_SAGA",
		Status:           domain.SAGA_PROCESSING,
		CurrentStepIndex: 0,
	}
	sagaHandler := registry.NewSagaHandler(orderSaga, s.repo, s.logger)
	reservedSeatStep := &domain.SagaStep{
		ID:                    uuid.New(),
		SagaID:                sagaId,
		Name:                  string(sagapkg.RESERVED_SEAT_STEP),
		Status:                domain.SAGA_STEP_PENDING,
		Order:                 0,
		ShouldPauseForPayment: false,
		Execute: func(ctx context.Context) error {
			_, err := s.eventClient.UpdateBatchSeatStatus(ctx, &eventv1.UpdateBatchSeatStatusRequest{
				SeatIds:   req.SeatIds,
				Status:    "reserved",
				BookingId: req.BookingId,
			})
			if err != nil {
				s.logger.Error("Saga reserve seat fail", zap.Error(err))
				return err
			}
			return nil
		},
		Compensate: func(ctx context.Context) error {
			_, err := s.eventClient.UpdateBatchSeatStatus(ctx, &eventv1.UpdateBatchSeatStatusRequest{
				SeatIds:   req.SeatIds,
				Status:    "available",
				BookingId: uuid.Nil.String(),
			})
			if err != nil {
				s.logger.Error("Saga undo-reserve seat fail", zap.Error(err))
				return err
			}
			return nil
		},
	}
	createPaymentStep := &domain.SagaStep{
		ID:                    uuid.New(),
		SagaID:                sagaId,
		Name:                  string(sagapkg.CREATE_PAYMENT_INTENT_STEP),
		Status:                domain.SAGA_STEP_PENDING,
		Order:                 1,
		ShouldPauseForPayment: true,
		Execute: func(ctx context.Context) error {
			paymentRes, err := s.paymentClient.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
				UserId:    req.UserId,
				BookingId: req.BookingId,
				Price:     int32(req.TotalCents) / 10, // Convert to USD
				Currency:  "usd",
				// PaymentMethod: ,
				UserEmail: "nhkhai2805@gmail.com", // update later
			})
			if err != nil {
				s.logger.Error("Saga payment step fail", zap.Error(err))
				return err
			}
			sagaHandler.SetPaymentResponse(paymentRes)
			return nil
		},
		Compensate: func(ctx context.Context) error {
			// write refund function
			return nil
		},
	}

	sagaHandler.AddStep(reservedSeatStep)
	sagaHandler.AddStep(createPaymentStep)
	// register more steps...
	// .
	// .
	// .
	err := s.repo.Create(ctx, sagaHandler.GetSaga())
	if err != nil {
		return nil, err
	}
	return sagaHandler, nil
}

type StartOrderSagaResponse struct {
	Status              string
	Message             string
	PaymentClientSecret string
	SagaId              string
}

func (s *SagaService) StartOrderSaga(ctx context.Context, req *StartOrderSagaRequest) (*StartOrderSagaResponse, error) {
	sagaHandler, err := s.InitializeSagaHandler(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fail to initialize saga order handler: %w", err)
	}
	err = sagaHandler.Execute(ctx, 0)
	if err != nil {
		return nil, err
	}
	return &StartOrderSagaResponse{
		Status: "success",
		SagaId: sagaHandler.GetSagaID().String(),
		// Message: sagaHandler.GetPaymentResponse(),
		PaymentClientSecret: sagaHandler.GetPaymentResponse().PaymentIntentClientSecret,
	}, nil
}

func (s *SagaService) HandleSagaAferPaymentSuccess(ctx context.Context, req json.RawMessage) error {
	// TODO
	var paymentEvent pkgtyped.PaymentEvent
	err := json.Unmarshal(req, &paymentEvent)
	if err != nil {
		s.logger.Sugar().Errorf("fail to unmarshal payment event", zap.Error(err))
		return fmt.Errorf("fail to unmarshal payment event: %w", err)
	}
	stripeEventData := paymentEvent.Data
	var paymentIntent stripe.PaymentIntent
	err = json.Unmarshal(stripeEventData.Raw, &paymentIntent)
	if err != nil {
		s.logger.Sugar().Errorf("fail to unmarshal payment intent", zap.Error(err))
		return fmt.Errorf("fail to unmarshal payment intent: %w", err)
	}

	// Update payment (status, payment_intent_id)
	s.logger.Info("Handling Saga after payment process successfully")
	return nil
}

func (s *SagaService) HandleSagaAfterPaymentFailure(ctx context.Context, req json.RawMessage) error {
	// TODO
	s.logger.Info("Handling Saga after payment process fail")
	return nil
}
