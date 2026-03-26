package grpc

import (
	"context"

	sagav1 "github.com/ticketbox/pkg/proto/saga/v1"
	"github.com/ticketbox/saga/internal/service"
	"go.uber.org/zap"
)

type SagaOrchestratorServer struct {
	sagav1.UnimplementedSagaOrchestratorServiceServer
	service *service.SagaService
	logger  *zap.Logger
}

func NewSagaOrchestratorServer(service *service.SagaService, logger *zap.Logger) *SagaOrchestratorServer {
	return &SagaOrchestratorServer{
		service: service,
		logger:  logger,
	}
}

func (s *SagaOrchestratorServer) StartOrderSaga(ctx context.Context, req *sagav1.StartOrderSagaRequest) (*sagav1.OrderSagaResponse, error) {
	res, err := s.service.StartOrderSaga(ctx, &service.StartOrderSagaRequest{
		BookingId:  req.BookingId,
		SeatIds:    req.SeatIds,
		UserId:     req.UserId,
		TotalCents: int(req.TotalCents),
	})
	if err != nil {
		return nil, err
	}

	return &sagav1.OrderSagaResponse{
		Status:                    res.Status,
		Message:                   res.Message,
		PaymentIntentClientSecret: res.PaymentClientSecret,
		SagaId:                    res.SagaId,
	}, nil
}

func (s *SagaOrchestratorServer) CompensateOrderSaga(ctx context.Context, req *sagav1.CompensateOrderSagaRequest) (*sagav1.OrderSagaResponse, error) {
	return nil, nil
}
