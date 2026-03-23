package kafka

import (
	"context"
	"fmt"

	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/saga/internal/service"
	"go.uber.org/zap"
)

type SagaOrchestratorConsumer struct {
	service            *service.SagaService
	logger             *zap.Logger
	sagaOrchesConsumer *pkgkafka.Consumer
}

func NewSagaOrchestratorConsumer(brokers []string, service *service.SagaService, logger *zap.Logger) *SagaOrchestratorConsumer {
	sagaOrchestratorConsumer := &SagaOrchestratorConsumer{
		service: service,
		logger:  logger,
	}
	sagaOrchestratorConsumer.sagaOrchesConsumer = pkgkafka.NewConsumer(
		brokers,
		"payment.events",
		"saga-payment-group",
		sagaOrchestratorConsumer.ConsumerHandler,
		logger)
	return sagaOrchestratorConsumer
}

func (s *SagaOrchestratorConsumer) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- s.sagaOrchesConsumer.Start(ctx)
	}()
	return <-errChan
}

func (s *SagaOrchestratorConsumer) ConsumerHandler(ctx context.Context, event pkgkafka.Event) error {
	switch event.Type {
	case "PaymentSucceed":
		return s.service.HandleSagaAferPaymentSuccess(ctx, event.Data)
	case "PaymentFail":
		return s.service.HandleSagaAfterPaymentFailure(ctx, event.Data)
	default:
		return fmt.Errorf("Event type %s cannot be handled", event.Type)
	}
}
