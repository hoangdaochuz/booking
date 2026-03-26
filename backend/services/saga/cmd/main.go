package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	"github.com/ticketbox/pkg/middleware"
	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	sagav1 "github.com/ticketbox/pkg/proto/saga/v1"
	saga_grpc "github.com/ticketbox/saga/internal/grpc"
	"github.com/ticketbox/saga/internal/kafka"
	"github.com/ticketbox/saga/internal/registry"
	"github.com/ticketbox/saga/internal/repository"
	"github.com/ticketbox/saga/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	sagaStepRegistry := registry.NewSagaStepRegistry()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("[SAGA Service]: cannot load config", zap.Error(err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("[SAGA Service]: Cannot create postgres poll", zap.Error(err))
	}
	// Connect to Event Service via gRPC
	eventConn, err := grpc.NewClient(cfg.EventServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to event service", zap.Error(err))
	}
	defer eventConn.Close()
	eventClient := eventv1.NewEventServiceClient(eventConn)

	paymentConn, err := grpc.NewClient(cfg.PaymentServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to payment service", zap.Error(err))
	}
	defer paymentConn.Close()
	paymentClient := paymentv1.NewPaymentServiceClient(paymentConn)

	bookingConn, err := grpc.NewClient(cfg.BookingServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to booking service", zap.Error(err))
	}
	bookingClient := bookingv1.NewBookingServiceClient(bookingConn)

	sagaRepository := repository.NewSagaRepository(pool)
	sagaService := service.NewSagaService(logger, sagaRepository, sagaStepRegistry, bookingClient, paymentClient, eventClient)
	// err = sagaService.RegisterSagaSteps(ctx)
	// if err != nil {
	// 	logger.Fatal("Fail to register saga step", zap.Error(err))
	// }

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)))
	sagaGrpcServer := saga_grpc.NewSagaOrchestratorServer(sagaService, logger)
	sagav1.RegisterSagaOrchestratorServiceServer(grpcServer, sagaGrpcServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	logger.Info("Saga service has started", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC serve failed", zap.Error(err))
		}
	}()

	sagaComsumer := kafka.NewSagaOrchestratorConsumer(cfg.KafkaBrokers, sagaService, logger)

	go func() {
		err := sagaComsumer.Start(ctx)
		if err != nil {
			logger.Fatal("Fail to start Saga consumer", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	logger.Info("Saga service is shutting down...")
	cancel()
	grpcServer.GracefulStop()
}
