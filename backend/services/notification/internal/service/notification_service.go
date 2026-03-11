package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ticketbox/notification/internal/domain"
	"github.com/ticketbox/notification/internal/repository"
)

type NotificationService struct {
	repo   repository.NotificationRepository
	sender NotificationSender
	logger *zap.Logger
}

func NewNotificationService(repo repository.NotificationRepository, sender NotificationSender, logger *zap.Logger) *NotificationService {
	return &NotificationService{repo: repo, sender: sender, logger: logger}
}

func (s *NotificationService) SendBookingConfirmation(ctx context.Context, data json.RawMessage) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	email, _ := payload["email"].(string)
	if email == "" {
		email = "unknown@ticketbox.com"
	}

	n := &domain.Notification{
		ID:        uuid.New(),
		Type:      "booking_confirmation",
		Recipient: email,
		Channel:   "email",
		Payload:   payload,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.Error("Failed to create notification", zap.Error(err))
		return err
	}

	if err := s.sender.Send(ctx, n); err != nil {
		s.logger.Error("Failed to send notification", zap.Error(err))
		return err
	}

	return s.repo.MarkSent(ctx, n.ID)
}

func (s *NotificationService) SendBookingFailure(ctx context.Context, data json.RawMessage) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	email, _ := payload["email"].(string)
	if email == "" {
		email = "unknown@ticketbox.com"
	}

	n := &domain.Notification{
		ID:        uuid.New(),
		Type:      "booking_failure",
		Recipient: email,
		Channel:   "email",
		Payload:   payload,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}
	if err := s.sender.Send(ctx, n); err != nil {
		return err
	}
	return s.repo.MarkSent(ctx, n.ID)
}

func (s *NotificationService) SendBookingCancellation(ctx context.Context, data json.RawMessage) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	email, _ := payload["email"].(string)
	if email == "" {
		email = "unknown@ticketbox.com"
	}

	n := &domain.Notification{
		ID:        uuid.New(),
		Type:      "booking_cancellation",
		Recipient: email,
		Channel:   "email",
		Payload:   payload,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}
	if err := s.sender.Send(ctx, n); err != nil {
		return err
	}
	return s.repo.MarkSent(ctx, n.ID)
}

func (s *NotificationService) SendWelcomeEmail(ctx context.Context, data json.RawMessage) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	email, _ := payload["email"].(string)
	if email == "" {
		email = "unknown@ticketbox.com"
	}

	n := &domain.Notification{
		ID:        uuid.New(),
		Type:      "welcome_email",
		Recipient: email,
		Channel:   "email",
		Payload:   payload,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}
	if err := s.sender.Send(ctx, n); err != nil {
		return err
	}
	return s.repo.MarkSent(ctx, n.ID)
}
