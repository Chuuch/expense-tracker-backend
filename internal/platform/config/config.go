package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:",squash"`
	Server ServerConfig `mapstructure:",squash"`
	DB     DBConfig     `mapstructure:",squash"`
	Auth   AuthConfig   `mapstructure:",squash"`
	Plaid  PlaidConfig  `mapstructure:",squash"`
	Redis  RedisConfig  `mapstructure:",squash"`
}

type AppConfig struct {
	AppEnv  string `mapstructure:"APP_ENV"`
	Version string `mapstructure:"APP_VERSION"`
}

type ServerConfig struct {
	Port           string        `mapstructure:"SERVER_PORT"`
	ReadTimeout    time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout   time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout    time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
	AllowedOrigins []string      `mapstructure:"SERVER_ALLOWED_ORIGINS"`
}

type DBConfig struct {
	URL             string        `mapstructure:"DB_URL"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

type AuthConfig struct {
	PasetoSymmetricKey string        `mapstructure:"AUTH_PASETO_KEY"` // 32-byte key
	AccessTokenTTL     time.Duration `mapstructure:"AUTH_ACCESS_TOKEN_TTL"`
	RefreshTokenTTL    time.Duration `mapstructure:"AUTH_REFRESH_TOKEN_TTL"`
	RateLimitRequests  int           `mapstructure:"AUTH_RATE_LIMIT_REQUESTS"`
	RateLimitWindow    time.Duration `mapstructure:"AUTH_RATE_LIMIT_WINDOW"`
}

type PlaidConfig struct{}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName(".env.development")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.SetDefault("AUTH_RATE_LIMIT_REQUESTS", 10)
	v.SetDefault("AUTH_RATE_LIMIT_WINDOW", "1m")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config file not found: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.Server.Port) == "" {
		return fmt.Errorf("invalid config: SERVER_PORT is required")
	}
	if strings.TrimSpace(c.DB.URL) == "" {
		return fmt.Errorf("invalid config: DB_URL is required")
	}
	if strings.TrimSpace(c.Auth.PasetoSymmetricKey) == "" {
		return fmt.Errorf("invalid config: AUTH_PASETO_KEY is required")
	}
	if len(c.Auth.PasetoSymmetricKey) != 32 {
		return fmt.Errorf("invalid config: AUTH_PASETO_KEY must be exactly 32 characters")
	}
	if c.Auth.AccessTokenTTL <= 0 {
		return fmt.Errorf("invalid config: AUTH_ACCESS_TOKEN_TTL must be greater than 0")
	}
	if c.Auth.RefreshTokenTTL <= 0 {
		return fmt.Errorf("invalid config: AUTH_REFRESH_TOKEN_TTL must be greater than 0")
	}
	if c.Auth.RateLimitRequests <= 0 {
		return fmt.Errorf("invalid config: AUTH_RATE_LIMIT_REQUESTS must be greater than 0")
	}
	if c.Auth.RateLimitWindow <= 0 {
		return fmt.Errorf("invalid config: AUTH_RATE_LIMIT_WINDOW must be greater than 0")
	}
	if strings.TrimSpace(c.Redis.Host) == "" {
		return fmt.Errorf("invalid config: REDIS_HOST is required")
	}
	if strings.TrimSpace(c.Redis.Port) == "" {
		return fmt.Errorf("invalid config: REDIS_PORT is required")
	}
	return nil
}
