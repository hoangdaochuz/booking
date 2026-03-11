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
	"google.golang.org/grpc/reflection"

	"github.com/ticketbox/event/internal/repository"
	"github.com/ticketbox/event/internal/service"
	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	"github.com/ticketbox/pkg/middleware"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"

	eventgrpc "github.com/ticketbox/event/internal/grpc"
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

	eventRepo := repository.NewPostgresEventRepository(pool)
	tierRepo := repository.NewPostgresTicketTierRepository(pool)
	eventService := service.NewEventService(eventRepo, tierRepo, logger)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	eventServer := eventgrpc.NewEventServer(eventService, logger)
	eventv1.RegisterEventServiceServer(grpcServer, eventServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down event service...")
		grpcServer.GracefulStop()
		cancel()
	}()

	logger.Info("Event service started", zap.String("port", cfg.GRPCPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("gRPC serve failed", zap.Error(err))
	}
}
