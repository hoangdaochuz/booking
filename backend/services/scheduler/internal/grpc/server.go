package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	schedulerv1 "github.com/ticketbox/pkg/proto/scheduler/v1"
	"github.com/ticketbox/scheduler/internal/domain"
	service_scheduler "github.com/ticketbox/scheduler/internal/service/scheduler"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SchedulerServiceServer struct {
	schedulerv1.UnimplementedSchedulerServiceServer
	service *service_scheduler.SchedulerService
	logger  *zap.Logger
}

func NewSchedulerServiceServer(service *service_scheduler.SchedulerService, logger *zap.Logger) *SchedulerServiceServer {
	return &SchedulerServiceServer{
		service: service,
		logger:  logger,
	}
}

func (s *SchedulerServiceServer) ListActiveSchedulers(ctx context.Context, req *schedulerv1.ListActiveSchedulersRequest) (*schedulerv1.ListActiveSchedulersResponse, error) {
	res, err := s.service.ListSchedulersConfig(ctx)
	if err != nil {
		s.logger.Sugar().Error("[Scheduler service][LisActiveSchedulers] Fail to get list schedulers config", zap.Error(err))
		return nil, err
	}
	schedulerCfgs := make([]*schedulerv1.SchedulerItem, 0, len(res))
	for _, cfg := range res {
		schedulerCfgs = append(schedulerCfgs, &schedulerv1.SchedulerItem{
			Id:             cfg.Id.String(),
			IsEnabled:      cfg.IsEnabled,
			Version:        int32(cfg.Version),
			Name:           cfg.Name,
			CronExpression: cfg.IntervalExpression,
			Timeout:        int32(cfg.Timeout / time.Second),
			CreatedAt:      timestamppb.New(cfg.CreatedAt),
			UpdatedAt:      timestamppb.New(cfg.UpdatedAt),
		})
	}
	return &schedulerv1.ListActiveSchedulersResponse{
		Schedulers: schedulerCfgs,
	}, nil
}

func (s *SchedulerServiceServer) UpdateSchedulerById(ctx context.Context, req *schedulerv1.UpdateSchedulerByIdRequest) (*emptypb.Empty, error) {
	schedulerCfgId, err := uuid.Parse(req.Id)
	if err != nil {
		s.logger.Sugar().Error("[Scheduler service][UpdateSchedulerById] Scheduler config id is invalid", zap.Error(err))
		return nil, fmt.Errorf("scheduler config id is invalid: %w", err)
	}
	err = s.service.UpdateSchedulerConfigById(ctx, schedulerCfgId, domain.SchedulerConfig{
		IsEnabled:          req.IsEnable,
		IntervalExpression: req.CronExpression,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
