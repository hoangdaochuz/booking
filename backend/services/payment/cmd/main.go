package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	payment_grpc "github.com/ticketbox/payment/internal/grpc"
	"github.com/ticketbox/payment/internal/repository"
	"github.com/ticketbox/payment/internal/service"
	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	"github.com/ticketbox/pkg/middleware"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	paymentRepo := repository.NewPostgresPaymentRepository(pool)
	paymentService := service.NewPaymentService(paymentRepo)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)))
	paymentServer := payment_grpc.NewPaymentServer(paymentService, logger)
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	logger.Info("Payment service has started", zap.String("port", cfg.GRPCPort))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC serve failed", zap.Error(err))
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("Payment service is shuting down....")
	cancel()
	grpcServer.GracefulStop()
}
