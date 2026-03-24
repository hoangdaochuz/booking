package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	stripe_go "github.com/stripe/stripe-go/v84"
	"github.com/ticketbox/payment/internal/gateway/stripe"
	payment_grpc "github.com/ticketbox/payment/internal/grpc"
	payment_http "github.com/ticketbox/payment/internal/http"
	"github.com/ticketbox/payment/internal/kafka"
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

	stripe_go.Key = cfg.StripeSecretKey

	paymentProducer := kafka.NewPaymentProducer(cfg.KafkaBrokers, logger)

	stripeGateway := stripe.NewStripePaymentGateway(logger, cfg.StripeSecretWebhook, paymentProducer)

	paymentRepo := repository.NewPostgresPaymentRepository(pool)
	paymentService := service.NewPaymentService(paymentRepo, logger, stripeGateway)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)))
	paymentServer := payment_grpc.NewPaymentServer(paymentService, logger)
	paymentv1.RegisterPaymentServiceServer(grpcServer, paymentServer)
	reflection.Register(grpcServer)

	// HTTP server for Webhook stripe/zalopay/momo payment gateway
	mux := http.NewServeMux()
	stripeWebhookGateway := stripe.NewStripePaymentGateway(logger, cfg.StripeSecretWebhook, paymentProducer)
	paymentHttpServerStripeHandler := payment_http.NewPaymentHttpHandler(paymentService, logger, stripeWebhookGateway)
	// Momo, zalopay,...

	mux.HandleFunc("/webhooks/stripe", paymentHttpServerStripeHandler.WebhookHandler)

	go func() {
		logger.Info("Payment http server has started on 8081 port")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			logger.Fatal("Fail to listen and serve payment http server", zap.Error(err))
		}
	}()

	// GRPC Server for internal communication
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
