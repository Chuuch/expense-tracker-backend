package config

import (
	"testing"
	"time"
)

func TestValidate_SucceedsForValidConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: ":8080"},
		DB:     DBConfig{URL: "postgres://user:pass@localhost:5432/db?sslmode=disable"},
		Auth: AuthConfig{
			PasetoSymmetricKey: "12345678901234567890123456789012",
			AccessTokenTTL:     15 * time.Minute,
			RefreshTokenTTL:    24 * time.Hour,
			RateLimitRequests:  10,
			RateLimitWindow:    time.Minute,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	if err := cfg.validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidate_FailsForInvalidPasetoKeyLength(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: ":8080"},
		DB:     DBConfig{URL: "postgres://user:pass@localhost:5432/db?sslmode=disable"},
		Auth: AuthConfig{
			PasetoSymmetricKey: "short-key",
			AccessTokenTTL:     15 * time.Minute,
			RefreshTokenTTL:    24 * time.Hour,
			RateLimitRequests:  10,
			RateLimitWindow:    time.Minute,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	if err := cfg.validate(); err == nil {
		t.Fatal("expected validation error for key length, got nil")
	}
}
