package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	authpostgres "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	authresend "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/resend"
	"github.com/chuuch/expense-tracker-backend/internal/auth/tasks"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	"github.com/chuuch/expense-tracker-backend/pkg/logger"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("worker: load config: %v", err)
	}

	if strings.TrimSpace(cfg.Resend.APIKey) == "" || strings.TrimSpace(cfg.Resend.From) == "" {
		log.Fatalf("worker: resend config is invalid")
	}

	logger.InitLogger(cfg.App.AppEnv)
	defer logger.Log.Sync()

	db, err := sql.Open("pgx", cfg.DB.URL)
	if err != nil {
		log.Fatalf("worker: open database: %v", err)
	}
	defer db.Close()

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()
	if err := db.PingContext(dbCtx); err != nil {
		logger.Log.Fatal("worker: database is not reachable", zap.Error(err))
	}

	q := postgresdb.New(db)
	userRepo := authpostgres.NewUserRepository(q)
	sender := authresend.NewVerificationEmailSEnder(cfg.Resend.APIKey, cfg.Resend.From)

	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{Concurrency: 5})
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeSendVerificationEmail, handleSendVerificationEmail(userRepo, sender))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := srv.Run(mux); err != nil {
			logger.Log.Fatal("worker: asynq run", zap.Error(err))
		}
	}()

	logger.Log.Info("worker: started")
	<-ctx.Done()
	logger.Log.Info("worker: shutting down")
	srv.Shutdown()
}

func handleSendVerificationEmail(userRepo interfaces.UserRepository, sender interfaces.VerificationEmailSender) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p tasks.SendVerificationEmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}
		user, err := userRepo.GetByID(ctx, p.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return nil
		}
		if user.VerificationCode == nil || user.VerificationCodeExpiresAt == nil {
			return nil
		}
		if time.Now().After(*user.VerificationCodeExpiresAt) {
			return nil
		}
		return sender.SendVerificationEmail(ctx, user.Email, *user.VerificationCode)
	}
}
