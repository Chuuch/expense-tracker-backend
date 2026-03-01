package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func BuildDependencies(cfg *config.Config, log *zap.Logger) (*Dependencies, error) {
	_ = log

	db, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()
	if err := db.PingContext(dbCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	redisCtx, redisCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer redisCancel()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return nil, err
	}

	q := postgresdb.New(db)

	authMod, err := buildAuthModule(cfg, q, redisClient)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return nil, err
	}

	expenseHandler := buildExpenseModule(q)

	return &Dependencies{
		DB:                   db,
		Redis:                redisClient,
		UserHandler:          authMod.handler,
		ExpenseHandler:       expenseHandler,
		TokenUsecase:         authMod.tokenUsecase,
		AccessTokenBlacklist: authMod.accessTokenBlacklist,
	}, nil
}
