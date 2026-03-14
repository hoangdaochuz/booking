package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	redis *redis.Client
}

type RedisLock struct {
	client *redis.Client
	ttl    time.Duration
	key    string
	token  string
}

var (
	ErrLockAlreadyAcquire error = errors.New("Lock cannot acquire")
	ErrLockNotOwned       error = errors.New("Lock not owned")
)

func NewClient(redisClient *redis.Client) *RedisClient {
	return &RedisClient{
		redis: redisClient,
	}
}

// Lua script for release lock
var releaseLockScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

var extendLockScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("PEXPIRE", KEYS[1] ,ARGV[2])
	else
		return 0
	end
`)

func (r *RedisClient) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*RedisLock, error) {
	token, err := r.generateToken()
	if err != nil {
		return nil, err
	}
	result, err := r.redis.SetArgs(ctx, key, token, redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	}).Result()
	if err != nil {
		return nil, err
	}

	if result != "OK" {
		return nil, ErrLockAlreadyAcquire
	}

	return &RedisLock{
		client: r.redis,
		ttl:    ttl,
		key:    key,
		token:  token,
	}, nil
}

func (r *RedisClient) AcquireLockWithRetry(ctx context.Context, key string, ttl time.Duration, interval time.Duration) (*RedisLock, error) {
	for {
		lock, err := r.AcquireLock(ctx, key, ttl)
		if lock != nil {
			return lock, err
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("fail to create acquire lock: %w", err)
		case <-time.After(interval):
			// retry
		}
	}
}

func (r *RedisLock) ReleaseLock(ctx context.Context) error {
	result, err := releaseLockScript.Run(ctx, r.client, []string{r.key}, r.token).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrLockNotOwned
	}
	return nil
}

func (r *RedisLock) ExtendLock(ctx context.Context) error {
	ms := r.ttl.Milliseconds()
	result, err := extendLockScript.Run(ctx, r.client, []string{r.key}, r.token, ms).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrLockNotOwned
	}
	return nil
}

func (r *RedisClient) generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("fail to create a token")
	}
	token := hex.EncodeToString(b)
	return token, nil
}

func (r *RedisLock) Key() string {
	return r.key
}

func (r *RedisLock) Token() string {
	return r.token
}
