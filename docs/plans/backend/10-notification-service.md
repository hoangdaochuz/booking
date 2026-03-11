# Task 10: Notification Service — Kafka Consumer

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a purely event-driven notification service that consumes Kafka events and logs simulated notifications.

**Files:**
- Create: `backend/services/notification/internal/domain/notification.go`
- Create: `backend/services/notification/internal/repository/notification_repository.go`
- Create: `backend/services/notification/internal/repository/postgres_notification_repository.go`
- Create: `backend/services/notification/internal/service/notification_service.go`
- Create: `backend/services/notification/internal/service/sender.go`
- Create: `backend/services/notification/internal/kafka/consumer.go`
- Modify: `backend/services/notification/cmd/main.go`

---

### Step 1: Create domain

`backend/services/notification/internal/domain/notification.go`:
```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type Notification struct {
    ID        uuid.UUID
    Type      string
    Recipient string
    Channel   string
    Payload   map[string]interface{}
    Status    string
    SentAt    *time.Time
    CreatedAt time.Time
}
```

### Step 2: Create repository interface and implementation

`backend/services/notification/internal/repository/notification_repository.go`:
```go
package repository

import (
    "context"

    "github.com/ticketbox/notification/internal/domain"
)

type NotificationRepository interface {
    Create(ctx context.Context, notification *domain.Notification) error
    MarkSent(ctx context.Context, id uuid.UUID) error
}
```

Implement with standard Postgres INSERT/UPDATE pattern.

### Step 3: Create pluggable sender interface

`backend/services/notification/internal/service/sender.go`:
```go
package service

import (
    "context"

    "github.com/ticketbox/notification/internal/domain"
    "go.uber.org/zap"
)

type NotificationSender interface {
    Send(ctx context.Context, notification *domain.Notification) error
}

// LogSender for MVP — logs to stdout
type LogSender struct {
    logger *zap.Logger
}

func NewLogSender(logger *zap.Logger) *LogSender {
    return &LogSender{logger: logger}
}

func (s *LogSender) Send(ctx context.Context, n *domain.Notification) error {
    s.logger.Info("NOTIFICATION SENT",
        zap.String("type", n.Type),
        zap.String("recipient", n.Recipient),
        zap.String("channel", n.Channel),
        zap.Any("payload", n.Payload),
    )
    return nil
}
```

### Step 4: Create notification service

`backend/services/notification/internal/service/notification_service.go`:
```go
package service

// Methods:
// - SendBookingConfirmation(ctx, data) — creates notification, calls sender, persists
// - SendBookingFailure(ctx, data)
// - SendBookingCancellation(ctx, data)
// - SendWelcomeEmail(ctx, data)
//
// Each method:
// 1. Unmarshal event data
// 2. Create Notification domain object
// 3. Persist to DB (outbox pattern — idempotency check)
// 4. Call sender.Send()
// 5. Mark as sent
```

### Step 5: Create Kafka consumer

`backend/services/notification/internal/kafka/consumer.go`:
```go
package kafka

import (
    "context"

    pkgkafka "github.com/ticketbox/pkg/kafka"
    "github.com/ticketbox/notification/internal/service"
    "go.uber.org/zap"
)

type NotificationConsumer struct {
    bookingConsumer *pkgkafka.Consumer
    userConsumer    *pkgkafka.Consumer
    service         *service.NotificationService
    logger          *zap.Logger
}

func NewNotificationConsumer(brokers []string, svc *service.NotificationService, logger *zap.Logger) *NotificationConsumer {
    nc := &NotificationConsumer{service: svc, logger: logger}

    nc.bookingConsumer = pkgkafka.NewConsumer(
        brokers, "booking.events", "notification-booking-group",
        nc.handleBookingEvent, logger,
    )
    nc.userConsumer = pkgkafka.NewConsumer(
        brokers, "user.events", "notification-user-group",
        nc.handleUserEvent, logger,
    )

    return nc
}

func (c *NotificationConsumer) Start(ctx context.Context) error {
    errCh := make(chan error, 2)
    go func() { errCh <- c.bookingConsumer.Start(ctx) }()
    go func() { errCh <- c.userConsumer.Start(ctx) }()
    return <-errCh
}

func (c *NotificationConsumer) handleBookingEvent(ctx context.Context, event pkgkafka.Event) error {
    switch event.Type {
    case "BookingConfirmed":
        return c.service.SendBookingConfirmation(ctx, event.Data)
    case "BookingFailed":
        return c.service.SendBookingFailure(ctx, event.Data)
    case "BookingCancelled":
        return c.service.SendBookingCancellation(ctx, event.Data)
    default:
        return nil
    }
}

func (c *NotificationConsumer) handleUserEvent(ctx context.Context, event pkgkafka.Event) error {
    switch event.Type {
    case "UserRegistered":
        return c.service.SendWelcomeEmail(ctx, event.Data)
    default:
        return nil
    }
}
```

### Step 6: Wire up main.go

`backend/services/notification/cmd/main.go`:

**No gRPC server.** Only:
1. Connect to Postgres
2. Create NotificationService with LogSender
3. Start Kafka consumers
4. Handle graceful shutdown

```go
func main() {
    // ... config, logger, db setup ...

    repo := repository.NewPostgresNotificationRepository(pool)
    sender := service.NewLogSender(logger)
    svc := service.NewNotificationService(repo, sender, logger)

    consumer := kafka.NewNotificationConsumer(cfg.KafkaBrokers, svc, logger)

    go func() {
        if err := consumer.Start(ctx); err != nil {
            logger.Fatal("Consumer failed", zap.Error(err))
        }
    }()

    // Wait for shutdown signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    cancel()
}
```

### Step 7: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/notification
go get github.com/google/uuid
go get github.com/jackc/pgx/v5
go get go.uber.org/zap
go mod tidy
go build ./...
```
Expected: No errors.

### Step 8: Commit

```bash
git add backend/services/notification/
git commit -m "feat(notification): add event-driven notification service"
```
