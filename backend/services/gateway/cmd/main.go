package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ticketbox/pkg/config"
	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	userv1 "github.com/ticketbox/pkg/proto/user/v1"

	"github.com/ticketbox/gateway/internal/router"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to gRPC services
	userConn, err := grpc.NewClient(cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16*1024*1024)), // 16MB
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(16*1024*1024)), // 16MB
	)
	if err != nil {
		logger.Fatal("Failed to connect to user service", zap.Error(err))
	}
	defer userConn.Close()

	eventConn, err := grpc.NewClient(cfg.EventServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16*1024*1024)), // 16MB
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(16*1024*1024)), // 16MB
	)
	if err != nil {
		logger.Fatal("Failed to connect to event service", zap.Error(err))
	}
	defer eventConn.Close()

	bookingConn, err := grpc.NewClient(cfg.BookingServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16*1024*1024)), // 16MB
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(16*1024*1024)), // 16MB
	)
	if err != nil {
		logger.Fatal("Failed to connect to booking service", zap.Error(err))
	}
	defer bookingConn.Close()

	paymentConn, err := grpc.NewClient(cfg.PaymentServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16*1024*1024)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(16*1024*1024)),
	)
	if err != nil {
		logger.Fatal("Failed to connect to payment service", zap.Error(err))
	}
	defer paymentConn.Close()

	userClient := userv1.NewUserServiceClient(userConn)
	eventClient := eventv1.NewEventServiceClient(eventConn)
	bookingClient := bookingv1.NewBookingServiceClient(bookingConn)
	paymentClient := paymentv1.NewPaymentServiceClient(paymentConn)
	// Connect to Redis
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal("Failed to parse Redis URL", zap.Error(err))
	}
	redisClient := redis.NewClient(opt)

	// Setup router
	r := router.SetupRouter(userClient, eventClient, bookingClient, paymentClient, redisClient)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: r,
	}

	go func() {
		logger.Info("Gateway started", zap.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
