package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/chuuch/expense-tracker-backend/internal/app"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/chuuch/expense-tracker-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Init config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Init logger
	logger.InitLogger(cfg.App.AppEnv)
	defer logger.Log.Sync()

	deps, err := app.BuildDependencies(cfg, logger.Log)
	if err != nil {
		logger.Log.Fatal("build dependencies failed", zap.Error(err))
	}
	defer deps.Close()

	// Init application
	application := app.NewApp(
		cfg,
		logger.Log,
		deps.UserHandler,
		deps.ExpenseHandler,
		deps.GoalHandler,
		deps.AnalyticsHandler,
		deps.TokenUsecase,
		deps.AccessTokenBlacklist,
		deps.DB,
		deps.Redis,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := application.Run(ctx); err != nil {
		logger.Log.Fatal("run application failed", zap.Error(err))
	}
}
