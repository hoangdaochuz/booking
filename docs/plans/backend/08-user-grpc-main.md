# Task 8: User Service — gRPC Server & Main

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire up gRPC server, Kafka event producer, and service entrypoint.

**Files:**
- Create: `backend/services/user/internal/grpc/server.go`
- Create: `backend/services/user/internal/kafka/producer.go`
- Modify: `backend/services/user/cmd/main.go`

---

### Step 1: Create gRPC server handler

`backend/services/user/internal/grpc/server.go`:
```go
package grpc

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "go.uber.org/zap"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"

    "github.com/ticketbox/user/internal/domain"
    "github.com/ticketbox/user/internal/service"
    userv1 "github.com/ticketbox/pkg/proto/user/v1"
)

type UserServer struct {
    userv1.UnimplementedUserServiceServer
    authService *service.AuthService
    producer    EventProducer
    logger      *zap.Logger
}

type EventProducer interface {
    PublishUserRegistered(ctx context.Context, user *domain.User) error
}

func NewUserServer(authService *service.AuthService, producer EventProducer, logger *zap.Logger) *UserServer {
    return &UserServer{authService: authService, producer: producer, logger: logger}
}

func (s *UserServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.AuthResponse, error) {
    user, accessToken, refreshToken, err := s.authService.Register(ctx, req.Email, req.Password, req.Name)
    if err != nil {
        if errors.Is(err, service.ErrEmailExists) {
            return nil, status.Error(codes.AlreadyExists, "email already registered")
        }
        s.logger.Error("Register failed", zap.Error(err))
        return nil, status.Error(codes.Internal, "registration failed")
    }

    if err := s.producer.PublishUserRegistered(ctx, user); err != nil {
        s.logger.Error("Failed to publish UserRegistered event", zap.Error(err))
    }

    return &userv1.AuthResponse{
        AccessToken: accessToken, RefreshToken: refreshToken, User: toUserProfile(user),
    }, nil
}

func (s *UserServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.AuthResponse, error) {
    user, accessToken, refreshToken, err := s.authService.Login(ctx, req.Email, req.Password)
    if err != nil {
        if errors.Is(err, service.ErrInvalidCredentials) {
            return nil, status.Error(codes.Unauthenticated, "invalid credentials")
        }
        s.logger.Error("Login failed", zap.Error(err))
        return nil, status.Error(codes.Internal, "login failed")
    }

    return &userv1.AuthResponse{
        AccessToken: accessToken, RefreshToken: refreshToken, User: toUserProfile(user),
    }, nil
}

func (s *UserServer) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.AuthResponse, error) {
    user, accessToken, refreshToken, err := s.authService.RefreshToken(ctx, req.RefreshToken)
    if err != nil {
        if errors.Is(err, service.ErrTokenExpired) {
            return nil, status.Error(codes.Unauthenticated, "token expired or revoked")
        }
        s.logger.Error("RefreshToken failed", zap.Error(err))
        return nil, status.Error(codes.Internal, "refresh failed")
    }

    return &userv1.AuthResponse{
        AccessToken: accessToken, RefreshToken: refreshToken, User: toUserProfile(user),
    }, nil
}

func (s *UserServer) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
    return &userv1.LogoutResponse{}, nil
}

func (s *UserServer) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.UserProfile, error) {
    userID, err := uuid.Parse(req.UserId)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid user ID")
    }

    user, err := s.authService.GetProfile(ctx, userID)
    if err != nil {
        return nil, status.Error(codes.NotFound, "user not found")
    }

    return toUserProfile(user), nil
}

func (s *UserServer) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UserProfile, error) {
    userID, err := uuid.Parse(req.UserId)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid user ID")
    }

    user, err := s.authService.UpdateProfile(ctx, userID, req.Name, req.Email)
    if err != nil {
        return nil, status.Error(codes.Internal, "update failed")
    }

    return toUserProfile(user), nil
}

func (s *UserServer) ValidateToken(ctx context.Context, req *userv1.ValidateTokenRequest) (*userv1.ValidateTokenResponse, error) {
    claims, err := s.authService.ValidateToken(req.AccessToken)
    if err != nil {
        return &userv1.ValidateTokenResponse{Valid: false}, nil
    }

    return &userv1.ValidateTokenResponse{
        Valid: true, UserId: claims.UserID, Email: claims.Email, Role: claims.Role,
    }, nil
}

func toUserProfile(user *domain.User) *userv1.UserProfile {
    return &userv1.UserProfile{
        Id: user.ID.String(), Email: user.Email, Name: user.Name,
        Role: user.Role, CreatedAt: timestamppb.New(user.CreatedAt),
    }
}
```

### Step 2: Create Kafka event producer

`backend/services/user/internal/kafka/producer.go`:
```go
package kafka

import (
    "context"
    "encoding/json"
    "time"

    pkgkafka "github.com/ticketbox/pkg/kafka"
    "github.com/ticketbox/user/internal/domain"
    "go.uber.org/zap"
)

const TopicUserEvents = "user.events"

type UserEventProducer struct {
    producer *pkgkafka.Producer
    logger   *zap.Logger
}

func NewUserEventProducer(brokers []string, logger *zap.Logger) *UserEventProducer {
    producer := pkgkafka.NewProducer(brokers, []string{TopicUserEvents}, logger)
    return &UserEventProducer{producer: producer, logger: logger}
}

func (p *UserEventProducer) PublishUserRegistered(ctx context.Context, user *domain.User) error {
    data, err := json.Marshal(map[string]interface{}{
        "user_id": user.ID.String(),
        "email":   user.Email,
        "name":    user.Name,
    })
    if err != nil {
        return err
    }

    event := pkgkafka.Event{
        Type:      "UserRegistered",
        Timestamp: time.Now(),
        Data:      data,
    }

    return p.producer.Publish(ctx, TopicUserEvents, user.ID.String(), event)
}

func (p *UserEventProducer) Close() error {
    return p.producer.Close()
}
```

### Step 3: Wire up main.go

`backend/services/user/cmd/main.go`:
```go
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
```

### Step 4: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/user
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get github.com/ticketbox/pkg
go mod tidy
go build ./...
```
Expected: No errors.

### Step 5: Commit

```bash
git add backend/services/user/
git commit -m "feat(user): add gRPC server, Kafka producer, and main entrypoint"
```
