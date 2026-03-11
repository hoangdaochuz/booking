# Task 6: User Service — Domain & Repository

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create user domain entities and Postgres repository with interface-based design.

**Files:**
- Create: `backend/services/user/internal/domain/user.go`
- Create: `backend/services/user/internal/repository/user_repository.go`
- Create: `backend/services/user/internal/repository/postgres_user_repository.go`

---

### Step 1: Create domain entities

`backend/services/user/internal/domain/user.go`:
```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
    Name         string
    Role         string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type RefreshToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    TokenHash string
    ExpiresAt time.Time
    RevokedAt *time.Time
    CreatedAt time.Time
}
```

### Step 2: Create repository interface

`backend/services/user/internal/repository/user_repository.go`:
```go
package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/ticketbox/user/internal/domain"
)

type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
}

type RefreshTokenRepository interface {
    Create(ctx context.Context, token *domain.RefreshToken) error
    GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
    Revoke(ctx context.Context, id uuid.UUID) error
    RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}
```

### Step 3: Implement Postgres repository

`backend/services/user/internal/repository/postgres_user_repository.go`:
```go
package repository

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/ticketbox/user/internal/domain"
)

var ErrNotFound = errors.New("not found")

type PostgresUserRepository struct {
    pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
    return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
    query := `INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
    _, err := r.pool.Exec(ctx, query,
        user.ID, user.Email, user.PasswordHash, user.Name, user.Role, user.CreatedAt, user.UpdatedAt)
    if err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE id = $1`
    user := &domain.User{}
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE email = $1`
    user := &domain.User{}
    err := r.pool.QueryRow(ctx, query, email).Scan(
        &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get user by email: %w", err)
    }
    return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
    query := `UPDATE users SET email = $2, name = $3, updated_at = $4 WHERE id = $1`
    _, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Name, time.Now())
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }
    return nil
}

// ─── Refresh Token Repository ───

type PostgresRefreshTokenRepository struct {
    pool *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(pool *pgxpool.Pool) *PostgresRefreshTokenRepository {
    return &PostgresRefreshTokenRepository{pool: pool}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
    query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
              VALUES ($1, $2, $3, $4, $5)`
    _, err := r.pool.Exec(ctx, query,
        token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
    if err != nil {
        return fmt.Errorf("create refresh token: %w", err)
    }
    return nil
}

func (r *PostgresRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
    query := `SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
              FROM refresh_tokens WHERE token_hash = $1`
    token := &domain.RefreshToken{}
    err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
        &token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.RevokedAt, &token.CreatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("get refresh token: %w", err)
    }
    return token, nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
    query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`
    _, err := r.pool.Exec(ctx, query, id)
    return err
}

func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
    query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
    _, err := r.pool.Exec(ctx, query, userID)
    return err
}
```

### Step 4: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/user
go get github.com/google/uuid
go get github.com/jackc/pgx/v5
go mod tidy
go build ./...
```
Expected: No errors.

### Step 5: Commit

```bash
git add backend/services/user/
git commit -m "feat(user): add domain entities and Postgres repository"
```
