package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	bookinggrpc "github.com/ticketbox/booking/internal/grpc"
	"github.com/ticketbox/booking/internal/repository"
	"github.com/ticketbox/booking/internal/service"
	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	"github.com/ticketbox/pkg/middleware"
	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Connect to Event Service via gRPC
	eventConn, err := grpc.NewClient(cfg.EventServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to event service", zap.Error(err))
	}
	defer eventConn.Close()
	eventClient := eventv1.NewEventServiceClient(eventConn)

	bookingRepo := repository.NewPostgresBookingRepository(pool)
	bookingService := service.NewBookingService(bookingRepo, eventClient, logger)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	bookingServer := bookinggrpc.NewBookingServer(bookingService, logger)
	bookingv1.RegisterBookingServiceServer(grpcServer, bookingServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down booking service...")
		grpcServer.GracefulStop()
		cancel()
	}()

	logger.Info("Booking service started", zap.String("port", cfg.GRPCPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("gRPC serve failed", zap.Error(err))
	}
}
