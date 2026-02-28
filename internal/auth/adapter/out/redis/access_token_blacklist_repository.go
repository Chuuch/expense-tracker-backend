package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type AccessTokenBlacklistRepository struct {
	client *goredis.Client
}

func NewAccessTokenBlacklistRepository(client *goredis.Client) *AccessTokenBlacklistRepository {
	return &AccessTokenBlacklistRepository{client: client}
}

func (r *AccessTokenBlacklistRepository) Add(ctx context.Context, token string, ttl time.Duration) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("blacklist add: token is required")
	}
	if ttl <= 0 {
		return fmt.Errorf("blacklist add: ttl must be > 0")
	}

	key := blacklistKey(token)
	if err := r.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("blacklist add: %w", err)
	}

	return nil
}

func (r *AccessTokenBlacklistRepository) Contains(ctx context.Context, token string) (bool, error) {
	if strings.TrimSpace(token) == "" {
		return false, nil
	}

	key := blacklistKey(token)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("blacklist contains: %w", err)
	}

	return exists > 0, nil
}

func blacklistKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "auth:blacklist:access:" + hex.EncodeToString(sum[:])
}
