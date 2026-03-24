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
	sagav1 "github.com/ticketbox/pkg/proto/saga/v1"
	saga_grpc "github.com/ticketbox/saga/internal/grpc"
	"github.com/ticketbox/saga/internal/kafka"
	"github.com/ticketbox/saga/internal/registry"
	"github.com/ticketbox/saga/internal/repository"
	"github.com/ticketbox/saga/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = registry.NewSagaStepRegistry()
	// sagaStepRegistry.Register("PaymentCreate", registry.SagaStepProcessor{
	// 	Execute: paymentClient.CreatePayment,
	// 	Compensate: paymentClient.Refund,
	// })
	// sagaStepRegistry.Register((""))
	// ....
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
	sagaRepository := repository.NewSagaRepository(pool)
	sagaService := service.NewSagaService(logger, sagaRepository)
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
