package service_scheduler

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ticketbox/pkg/outbound"
	"github.com/ticketbox/pkg/topics"
	"github.com/ticketbox/scheduler/internal/domain"
	cronjob "github.com/ticketbox/scheduler/internal/jobs"
	"github.com/ticketbox/scheduler/internal/repository"
	"go.uber.org/zap"
)

type SchedulerService struct {
	logger              *zap.Logger
	schedulerConfigRepo repository.SchedulerRepository
	outboundEventRepo   repository.OutboundEventRepository
	uow                 repository.UnitOfWork
	cronjobManager      *cronjob.CronJobManager
}

func NewSchedulerService(logger *zap.Logger, schedulerConfigRepo repository.SchedulerRepository, outboundEventRepo repository.OutboundEventRepository, uow repository.UnitOfWork, cronjobManager *cronjob.CronJobManager) *SchedulerService {
	return &SchedulerService{
		logger:              logger,
		schedulerConfigRepo: schedulerConfigRepo,
		outboundEventRepo:   outboundEventRepo,
		uow:                 uow,
		cronjobManager:      cronjobManager,
	}
}

func (s *SchedulerService) ListSchedulersConfig(ctx context.Context) ([]domain.SchedulerConfig, error) {
	return s.schedulerConfigRepo.ListSchedulersConfig(ctx)
}

func (s *SchedulerService) UpdateSchedulerConfigById(ctx context.Context, id uuid.UUID, target domain.SchedulerConfig) error {
	return s.uow.Execute(ctx, func(repos *repository.UowRepos) error {
		err := repos.SchedulerConfigRepo().UpdateById(ctx, id, target)
		if err != nil {
			s.logger.Sugar().Error("[SchedulerService][UpdateSchedulerConfigById] Fail to update scheduler config", zap.Error(err))
			return err
		}

		payload, err := json.Marshal(target)
		if err != nil {
			s.logger.Sugar().Error("[SchedulerService][UpdateSchedulerConfigById] Fail to marshal target scheduler config", zap.Error(err))
			return err
		}

		err = repos.OutboundRepo().Create(ctx, outbound.OutboundEvent{
			Topic:     topics.TopicSchedulerConfigEvent,
			EventType: topics.TypeSchedulerConfigChanged,
			Payload:   payload,
			Status:    outbound.OutboundStatusPending,
		})
		if err != nil {
			s.logger.Sugar().Error("[SchedulerService][UpdateSchedulerConfigById] Fail to create a outbound event", zap.Error(err))
			return err
		}
		return nil
	})
}

func (s *SchedulerService) HandleSchedulerConfigChanged(ctx context.Context, payload domain.SchedulerConfig) error {
	return s.cronjobManager.ReRegisterJob(ctx, payload)
}
