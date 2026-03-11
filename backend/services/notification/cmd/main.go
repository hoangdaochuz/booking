package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	notifkafka "github.com/ticketbox/notification/internal/kafka"
	"github.com/ticketbox/notification/internal/repository"
	"github.com/ticketbox/notification/internal/service"
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

	repo := repository.NewPostgresNotificationRepository(pool)
	sender := service.NewLogSender(logger)
	svc := service.NewNotificationService(repo, sender, logger)

	consumer := notifkafka.NewNotificationConsumer(cfg.KafkaBrokers, svc, logger)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			logger.Fatal("Consumer failed", zap.Error(err))
		}
	}()

	logger.Info("Notification service started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down notification service...")
	cancel()
}
