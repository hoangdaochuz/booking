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
	return nil, nil
}

func (s *SagaOrchestratorServer) CompensateOrderSaga(ctx context.Context, req *sagav1.CompensateOrderSagaRequest) (*sagav1.OrderSagaResponse, error) {
	return nil, nil
}
