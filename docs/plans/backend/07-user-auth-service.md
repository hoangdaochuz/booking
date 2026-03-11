# Task 7: User Service — Business Logic & JWT

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement JWT token management and auth business logic (register, login, refresh, validate).

**Files:**
- Create: `backend/services/user/internal/service/jwt.go`
- Create: `backend/services/user/internal/service/auth_service.go`

---

### Step 1: Create JWT helper

`backend/services/user/internal/service/jwt.go`:
```go
package service

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type TokenClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

type JWTManager struct {
    secret         []byte
    accessTokenTTL time.Duration
}

func NewJWTManager(secret string, accessTokenTTLMinutes int) *JWTManager {
    return &JWTManager{
        secret:         []byte(secret),
        accessTokenTTL: time.Duration(accessTokenTTLMinutes) * time.Minute,
    }
}

func (j *JWTManager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
    claims := TokenClaims{
        UserID: userID.String(),
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "ticketbox",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secret)
}

func (j *JWTManager) ValidateToken(tokenStr string) (*TokenClaims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return j.secret, nil
    })
    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    claims, ok := token.Claims.(*TokenClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }

    return claims, nil
}

func GenerateRefreshToken() string {
    return uuid.New().String()
}

func HashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

### Step 2: Create auth service

`backend/services/user/internal/service/auth_service.go`:
```go
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
```

### Step 3: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/user
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get go.uber.org/zap
go mod tidy
go build ./...
```
Expected: No errors.

### Step 4: Commit

```bash
git add backend/services/user/
git commit -m "feat(user): add auth service with JWT and bcrypt"
```
