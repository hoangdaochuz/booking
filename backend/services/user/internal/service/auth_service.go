package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/ticketbox/user/internal/domain"
	"github.com/ticketbox/user/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrTokenExpired       = errors.New("token expired or revoked")
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *JWTManager
	refreshTokenTTL  time.Duration
	logger           *zap.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *JWTManager,
	refreshTokenTTLDays int,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		refreshTokenTTL:  time.Duration(refreshTokenTTLDays) * 24 * time.Hour,
		logger:           logger,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, string, string, error) {
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return nil, "", "", ErrEmailExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", "", fmt.Errorf("check email: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", fmt.Errorf("create user: %w", err)
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*domain.User, string, string, error) {
	tokenHash := HashToken(refreshTokenStr)

	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", "", ErrTokenExpired
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("get refresh token: %w", err)
	}

	if storedToken.RevokedAt != nil || time.Now().After(storedToken.ExpiresAt) {
		return nil, "", "", ErrTokenExpired
	}

	if err := s.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return nil, "", "", fmt.Errorf("revoke old token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, "", "", fmt.Errorf("get user: %w", err)
	}

	accessToken, newRefreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, newRefreshToken, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*TokenClaims, error) {
	return s.jwtManager.ValidateToken(tokenStr)
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, name, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) generateTokens(ctx context.Context, user *domain.User) (string, string, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshTokenStr := GenerateRefreshToken()
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: HashToken(refreshTokenStr),
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}

	return accessToken, refreshTokenStr, nil
}
