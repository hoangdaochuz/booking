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

func (s *SagaService) UpdateBatchSeatStatus(ctx context.Context, req *eventv1.UpdateBatchSeatStatusRequest) error {
	_, err := s.eventClient.UpdateBatchSeatStatus(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (s *SagaService) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	paymentRes, err := s.paymentClient.CreatePayment(ctx, req)
	return paymentRes, err
}

func (s *SagaService) RefundPayment(ctx context.Context) error {
	s.logger.Info("Implementing refund processing")
	return nil
}

func (s *SagaService) UpdateBookingStatus(ctx context.Context, req *bookingv1.UpdateBookingStatusByIdReq) error {
	_, err := s.bookingClient.UpdateBookingStatusById(ctx, req)
	return err
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
	reservedSeatProcessor := s.sagaStepRegistry.Get(string(sagapkg.RESERVED_SEAT_STEP))
	if reservedSeatProcessor == nil {
		return nil, fmt.Errorf("fail to get processor from saga step registry for step %s: ", sagapkg.RESERVED_SEAT_STEP)
	}

	paymentProcessor := s.sagaStepRegistry.Get(string(sagapkg.CREATE_PAYMENT_INTENT_STEP))
	if paymentProcessor == nil {
		return nil, fmt.Errorf("fail to get processor from saga step registry for step %s: ", sagapkg.CREATE_PAYMENT_INTENT_STEP)
	}

	updateSeatAfterPaymentProcessor := s.sagaStepRegistry.Get(string(sagapkg.UPDATE_SEAT_BOOKED))
	if updateSeatAfterPaymentProcessor == nil {
		return nil, fmt.Errorf("fail to get processor from saga step registry for step %s", sagapkg.UPDATE_SEAT_BOOKED)
	}

	updateBookingStatusProcessor := s.sagaStepRegistry.Get(string(sagapkg.UPDATE_BOOKING_STATUS_CONFIRMED))
	if updateBookingStatusProcessor == nil {
		return nil, fmt.Errorf("fail to get processor from saga step registry for step %s", sagapkg.UPDATE_BOOKING_STATUS_CONFIRMED)
	}
	reservedSeatStep := &domain.SagaStep{
		ID:                    uuid.New(),
		SagaID:                sagaId,
		Name:                  string(sagapkg.RESERVED_SEAT_STEP),
		Status:                domain.SAGA_STEP_PENDING,
		Order:                 0,
		ShouldPauseForPayment: false,
		Execute: func(ctx context.Context) error {
			err := reservedSeatProcessor.Execute.(func(ctx context.Context, req *eventv1.UpdateBatchSeatStatusRequest) error)(ctx, &eventv1.UpdateBatchSeatStatusRequest{
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
			err := reservedSeatProcessor.Compensate.(func(ctx context.Context, req *eventv1.UpdateBatchSeatStatusRequest) error)(ctx, &eventv1.UpdateBatchSeatStatusRequest{
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
			paymentRes, err := paymentProcessor.Execute.(func(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error))(ctx, &paymentv1.CreatePaymentRequest{
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
			err := paymentProcessor.Compensate.(func(ctx context.Context) error)(ctx)
			if err != nil {
				s.logger.Sugar().Error("saga payment refund step fail", zap.Error(err))
				return err
			}
			return nil
		},
	}

	updateSeatAfterPaymentStep := &domain.SagaStep{
		ID:                    uuid.New(),
		SagaID:                sagaId,
		Name:                  string(sagapkg.UPDATE_SEAT_BOOKED),
		Status:                domain.SAGA_STEP_PENDING,
		Order:                 2,
		ShouldPauseForPayment: false,
		Execute: func(ctx context.Context) error {
			err := reservedSeatProcessor.Execute.(func(ctx context.Context, req *eventv1.UpdateBatchSeatStatusRequest) error)(ctx, &eventv1.UpdateBatchSeatStatusRequest{
				SeatIds:   req.SeatIds,
				Status:    "booked",
				BookingId: req.BookingId,
			})
			if err != nil {
				s.logger.Error("Saga update seat to booked fail", zap.Error(err))
				return err
			}
			return nil
		},
		Compensate: func(ctx context.Context) error {
			err := reservedSeatProcessor.Execute.(func(ctx context.Context, req *eventv1.UpdateBatchSeatStatusRequest) error)(ctx, &eventv1.UpdateBatchSeatStatusRequest{
				SeatIds:   req.SeatIds,
				Status:    "available",
				BookingId: req.BookingId,
			})
			if err != nil {
				s.logger.Error("Saga update seat to booked fail", zap.Error(err))
				return err
			}
			return nil
		},
	}

	updateBookingToConfirmed := &domain.SagaStep{
		ID:                    uuid.New(),
		SagaID:                sagaId,
		Name:                  string(sagapkg.UPDATE_BOOKING_STATUS_CONFIRMED),
		Status:                domain.SAGA_STEP_PENDING,
		Order:                 3,
		ShouldPauseForPayment: false,
		Execute: func(ctx context.Context) error {
			err := updateBookingStatusProcessor.Execute.(func(ctx context.Context, req *bookingv1.UpdateBookingStatusByIdReq) error)(ctx, &bookingv1.UpdateBookingStatusByIdReq{
				Id:     req.BookingId,
				Status: "CONFIRMED",
			})
			if err != nil {
				s.logger.Error("fail to update booking status to confirmed after payment", zap.Error(err))
			}
			return nil
		},
		Compensate: func(ctx context.Context) error {
			err := updateBookingStatusProcessor.Execute.(func(ctx context.Context, req *bookingv1.UpdateBookingStatusByIdReq) error)(ctx, &bookingv1.UpdateBookingStatusByIdReq{
				Id:     req.BookingId,
				Status: "FAILED",
			})
			if err != nil {
				s.logger.Error("fail to update booking status to failed after payment", zap.Error(err))
			}
			return nil
		},
	}

	sagaHandler.AddStep(reservedSeatStep)
	sagaHandler.AddStep(createPaymentStep)
	sagaHandler.AddStep(updateSeatAfterPaymentStep)
	// sagaHandler.AddStep(reduceAvailableSeatNumber)
	sagaHandler.AddStep(updateBookingToConfirmed)
	// sagaHanlder.AddStep(sendMailConfirmToUser)
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
		Status:              "success",
		SagaId:              sagaHandler.GetSagaID().String(),
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

	_, err = s.paymentClient.UpdatePaymentStatusByPaymentIntentId(ctx, &paymentv1.UpdatePaymentStatusByPaymentIntentIdReq{
		PaymentIntentId: paymentEvent.Id,
		Status:          "success",
	})
	if err != nil {
		s.logger.Error("fail to update status of payment to success", zap.Error(err))
		return fmt.Errorf("fail to update status of payment to success: %w", err)
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
