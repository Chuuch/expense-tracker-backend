package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/chuuch/expense-tracker-backend/internal/app"
	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	authredis "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/redis"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	"github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/token/paseto"
	authusecase "github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/chuuch/expense-tracker-backend/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	// Init confg
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init logger
	logger.InitLogger(cfg.App.AppEnv)
	defer logger.Log.Sync()

	// Init database
	db, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		log.Fatalf("Open database connection failed: %v", err)
	}
	defer db.Close()

	// Init DI
	q := postgresdb.New(db)
	userRepo := postgres.NewUserRepository(q)

	// Init token usecase
	tokenUseCase, err := paseto.NewPasetoUsecase(cfg.Auth.PasetoSymmetricKey)
	if err != nil {
		log.Fatalf("paseto usecase initialization failed: %v", err)
	}

	refreshRepo := postgres.NewRefreshTokenRepository(q)

	userUsecase := authusecase.NewUserUsecase(userRepo)
	refreshUsecase := authusecase.NewRefreshTokenUsecase(
		refreshRepo,
		userRepo,
		tokenUseCase,
		cfg,
	)

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	accessTokenBlacklist := authredis.NewAccessTokenBlacklistRepository(redisClient)

	// Init user handler
	userHandler := authhttp.NewUserHandlerWithRefresh(userUsecase, tokenUseCase, refreshUsecase, cfg, accessTokenBlacklist)

	// Init application
	application := app.NewApp(cfg, logger.Log, userHandler, tokenUseCase, accessTokenBlacklist)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run application failed: %v", err)
	}
}
