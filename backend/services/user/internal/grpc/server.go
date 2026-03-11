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
