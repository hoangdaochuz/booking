package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
	"github.com/ticketbox/pkg/middleware"
	"github.com/ticketbox/pkg/outbound"
	schedulerv1 "github.com/ticketbox/pkg/proto/scheduler/v1"
	redis_pkg "github.com/ticketbox/pkg/redis"
	schedulergrpc "github.com/ticketbox/scheduler/internal/grpc"
	cronjob "github.com/ticketbox/scheduler/internal/jobs"
	"github.com/ticketbox/scheduler/internal/kafka"
	"github.com/ticketbox/scheduler/internal/repository"
	service_outbound "github.com/ticketbox/scheduler/internal/service/outbound"
	service_scheduler "github.com/ticketbox/scheduler/internal/service/scheduler"
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
	defer cancel()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	redisCln := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	_ = redis_pkg.NewClient(redisCln)

	schedulerCfgRepo := repository.NewSchedulerConfigRepo(pool, nil)
	outboundEventRepo := repository.NewOutboundEventRepository(pool, nil)
	uow := repository.NewUnitOfWork(pool)

	cronManager := cronjob.NewCronjobManager(logger, schedulerCfgRepo, cronjob.WithSecond, cronjob.WithCronjobErrHandler)

	schedulerSvc := service_scheduler.NewSchedulerService(logger, schedulerCfgRepo, outboundEventRepo, uow, cronManager)
	outboundEventProducer := kafka.NewOutboundEventProducer(logger, cfg.KafkaBrokers)
	outboundSvc := service_outbound.NewOutboundEventService(logger, outboundEventRepo, outboundEventProducer)
	schedulerServer := schedulergrpc.NewSchedulerServiceServer(schedulerSvc, logger)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	schedulerv1.RegisterSchedulerServiceServer(grpcServer, schedulerServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Fail to listen", zap.Error(err))
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down scheduler service...")
		grpcServer.GracefulStop()
		if err := cronManager.Stop(); err != nil {
			logger.Error("Fail to stop cron manager", zap.Error(err))
		}
		cancel()
	}()

	schedulerConfigConsumer := kafka.NewSchedulerConfigConsumer(cfg.KafkaBrokers, schedulerSvc, logger)

	go func() {
		err := schedulerConfigConsumer.Start(ctx)
		if err != nil {
			logger.Sugar().Error("OutboundEventConsumer started fail", zap.Error(err))
		}
	}()

	// OutboundEventPollerAndPublisher
	outboundEventHandler := outbound.NewOutboundHandler(
		logger,
		outboundEventProducer.Producer(),
		outbound.OutboundCfg{
			Interval: "10s",
		},
		outboundSvc.MarkOutboundEventsAsPublished,
		outboundSvc.ListPendingOutboundEvents,
	)

	err = outboundEventHandler.Start(ctx)
	if err != nil {
		logger.Error("Fail to start OutboundEventHandler", zap.Error(err))
	}

	// Cronjob init is synchronous on purpose: robfig cron Start() is already
	// non-blocking, so no extra goroutine is needed. Fail fast here so the
	// service never looks healthy while jobs are silently dead.
	if err := cronManager.LoadJobSchedulerConfigs(ctx); err != nil {
		logger.Fatal("Fail to load job scheduler configs", zap.Error(err))
	}
	// Register job
	reservationCleanerJob := cronjob.NewReservationCleanerJob()
	if err := cronManager.RegisterJob(ctx, reservationCleanerJob); err != nil {
		logger.Fatal("Fail to register reservation cleaner job", zap.Error(err))
	}
	// .... continue register here

	cronManager.Start()
	logger.Info("Scheduler service started", zap.String("port", cfg.GRPCPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("gRPC server failed", zap.Error(err))
	}
}
