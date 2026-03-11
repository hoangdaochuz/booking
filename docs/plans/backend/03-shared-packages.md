# Task 3: Shared Packages — Config, Database, Kafka

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create reusable shared libraries for configuration, database connections, Kafka producer/consumer, and gRPC middleware.

**Files:**
- Create: `backend/pkg/config/config.go`
- Create: `backend/pkg/database/postgres.go`
- Create: `backend/pkg/kafka/producer.go`
- Create: `backend/pkg/kafka/consumer.go`
- Create: `backend/pkg/middleware/logging.go`

---

### Step 1: Create config package

`backend/pkg/config/config.go`:
```go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    ServiceName string `mapstructure:"SERVICE_NAME"`
    GRPCPort    string `mapstructure:"GRPC_PORT"`
    HTTPPort    string `mapstructure:"HTTP_PORT"`

    DatabaseURL string `mapstructure:"DATABASE_URL"`
    RedisURL    string `mapstructure:"REDIS_URL"`

    KafkaBrokers []string `mapstructure:"KAFKA_BROKERS"`

    JWTSecret          string `mapstructure:"JWT_SECRET"`
    JWTAccessTokenTTL  int    `mapstructure:"JWT_ACCESS_TOKEN_TTL"`
    JWTRefreshTokenTTL int    `mapstructure:"JWT_REFRESH_TOKEN_TTL"`

    BookingMode string `mapstructure:"BOOKING_MODE"`
}

func Load() (*Config, error) {
    viper.AutomaticEnv()

    viper.SetDefault("GRPC_PORT", "50051")
    viper.SetDefault("HTTP_PORT", "8000")
    viper.SetDefault("REDIS_URL", "redis://localhost:6379")
    viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
    viper.SetDefault("JWT_ACCESS_TOKEN_TTL", 15)
    viper.SetDefault("JWT_REFRESH_TOKEN_TTL", 7)
    viper.SetDefault("BOOKING_MODE", "pessimistic")

    cfg := &Config{}
    cfg.ServiceName = viper.GetString("SERVICE_NAME")
    cfg.GRPCPort = viper.GetString("GRPC_PORT")
    cfg.HTTPPort = viper.GetString("HTTP_PORT")
    cfg.DatabaseURL = viper.GetString("DATABASE_URL")
    cfg.RedisURL = viper.GetString("REDIS_URL")
    cfg.KafkaBrokers = viper.GetStringSlice("KAFKA_BROKERS")
    cfg.JWTSecret = viper.GetString("JWT_SECRET")
    cfg.JWTAccessTokenTTL = viper.GetInt("JWT_ACCESS_TOKEN_TTL")
    cfg.JWTRefreshTokenTTL = viper.GetInt("JWT_REFRESH_TOKEN_TTL")
    cfg.BookingMode = viper.GetString("BOOKING_MODE")

    return cfg, nil
}
```

### Step 2: Create database package

`backend/pkg/database/postgres.go`:
```go
package database

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

func NewPostgresPool(ctx context.Context, databaseURL string, logger *zap.Logger) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, fmt.Errorf("parse database URL: %w", err)
    }

    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = 30 * time.Minute
    config.MaxConnIdleTime = 5 * time.Minute

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("create connection pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    logger.Info("Connected to PostgreSQL", zap.String("url", databaseURL))
    return pool, nil
}
```

### Step 3: Create Kafka producer

`backend/pkg/kafka/producer.go`:
```go
package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    kafkago "github.com/segmentio/kafka-go"
    "go.uber.org/zap"
)

type Event struct {
    Type      string          `json:"type"`
    Timestamp time.Time       `json:"timestamp"`
    Data      json.RawMessage `json:"data"`
}

type Producer struct {
    writers map[string]*kafkago.Writer
    logger  *zap.Logger
}

func NewProducer(brokers []string, topics []string, logger *zap.Logger) *Producer {
    writers := make(map[string]*kafkago.Writer)
    for _, topic := range topics {
        writers[topic] = &kafkago.Writer{
            Addr:         kafkago.TCP(brokers...),
            Topic:        topic,
            Balancer:     &kafkago.LeastBytes{},
            BatchTimeout: 10 * time.Millisecond,
            RequiredAcks: kafkago.RequireOne,
        }
    }
    return &Producer{writers: writers, logger: logger}
}

func (p *Producer) Publish(ctx context.Context, topic string, key string, event Event) error {
    writer, ok := p.writers[topic]
    if !ok {
        p.logger.Error("Unknown topic", zap.String("topic", topic))
        return fmt.Errorf("unknown topic: %s", topic)
    }

    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }

    msg := kafkago.Message{
        Key:   []byte(key),
        Value: data,
    }

    if err := writer.WriteMessages(ctx, msg); err != nil {
        p.logger.Error("Failed to publish", zap.String("topic", topic), zap.Error(err))
        return fmt.Errorf("publish to %s: %w", topic, err)
    }

    p.logger.Debug("Published event", zap.String("topic", topic), zap.String("type", event.Type))
    return nil
}

func (p *Producer) Close() error {
    for _, w := range p.writers {
        if err := w.Close(); err != nil {
            p.logger.Error("Failed to close writer", zap.Error(err))
        }
    }
    return nil
}
```

### Step 4: Create Kafka consumer

`backend/pkg/kafka/consumer.go`:
```go
package kafka

import (
    "context"
    "encoding/json"

    kafkago "github.com/segmentio/kafka-go"
    "go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, event Event) error

type Consumer struct {
    reader  *kafkago.Reader
    logger  *zap.Logger
    handler MessageHandler
}

func NewConsumer(brokers []string, topic string, groupID string, handler MessageHandler, logger *zap.Logger) *Consumer {
    reader := kafkago.NewReader(kafkago.ReaderConfig{
        Brokers:  brokers,
        Topic:    topic,
        GroupID:  groupID,
        MinBytes: 1,
        MaxBytes: 10e6,
    })

    return &Consumer{reader: reader, logger: logger, handler: handler}
}

func (c *Consumer) Start(ctx context.Context) error {
    c.logger.Info("Consumer started", zap.String("topic", c.reader.Config().Topic))

    for {
        select {
        case <-ctx.Done():
            return c.reader.Close()
        default:
            msg, err := c.reader.ReadMessage(ctx)
            if err != nil {
                if ctx.Err() != nil {
                    return nil
                }
                c.logger.Error("Read message failed", zap.Error(err))
                continue
            }

            var event Event
            if err := json.Unmarshal(msg.Value, &event); err != nil {
                c.logger.Error("Unmarshal event failed", zap.Error(err))
                continue
            }

            if err := c.handler(ctx, event); err != nil {
                c.logger.Error("Handle event failed",
                    zap.String("type", event.Type), zap.Error(err))
            }
        }
    }
}
```

### Step 5: Create logging middleware

`backend/pkg/middleware/logging.go`:
```go
package middleware

import (
    "context"
    "time"

    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
)

func UnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        duration := time.Since(start)

        st, _ := status.FromError(err)
        logger.Info("gRPC call",
            zap.String("method", info.FullMethod),
            zap.Duration("duration", duration),
            zap.String("status", st.Code().String()),
        )

        return resp, err
    }
}
```

### Step 6: Install dependencies

```bash
cd /Users/dev/work/booking/backend/pkg
go get github.com/spf13/viper
go get github.com/jackc/pgx/v5
go get github.com/segmentio/kafka-go
go get github.com/redis/go-redis/v9
go get go.uber.org/zap
go get google.golang.org/grpc
go mod tidy
```

### Step 7: Verify package compiles

```bash
cd /Users/dev/work/booking/backend/pkg
go build ./...
```
Expected: No errors.

### Step 8: Commit

```bash
git add backend/pkg/
git commit -m "feat(backend): add shared packages for config, database, kafka, middleware"
```
