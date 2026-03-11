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

	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	"github.com/ticketbox/pkg/middleware"
	userv1 "github.com/ticketbox/pkg/proto/user/v1"
	usergrpc "github.com/ticketbox/user/internal/grpc"
	userkafka "github.com/ticketbox/user/internal/kafka"
	"github.com/ticketbox/user/internal/repository"
	"github.com/ticketbox/user/internal/service"
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

	userRepo := repository.NewPostgresUserRepository(pool)
	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(pool)
	jwtManager := service.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessTokenTTL)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtManager, cfg.JWTRefreshTokenTTL, logger)

	producer := userkafka.NewUserEventProducer(cfg.KafkaBrokers, logger)
	defer producer.Close()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	userServer := usergrpc.NewUserServer(authService, producer, logger)
	userv1.RegisterUserServiceServer(grpcServer, userServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down user service...")
		grpcServer.GracefulStop()
		cancel()
	}()

	logger.Info("User service started", zap.String("port", cfg.GRPCPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("gRPC serve failed", zap.Error(err))
	}
}
