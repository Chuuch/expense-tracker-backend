package app

import (
	"database/sql"

	analyticshttp "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/in/http"
	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	expensehttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	goalhttp "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/in/http"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	DB                   *sql.DB
	Redis                *redis.Client
	UserHandler          *authhttp.UserHandler
	ExpenseHandler       *expensehttp.ExpenseHandler
	GoalHandler          *goalhttp.GoalHandler
	AnalyticsHandler     *analyticshttp.AnalyticsHandler
	TokenUsecase         interfaces.TokenUsecase
	AccessTokenBlacklist interfaces.AccessTokenBlacklist
	AsynqClient          *asynq.Client
}

func (d *Dependencies) Close() {
	if d.Redis != nil {
		_ = d.Redis.Close()
	}
	if d.DB != nil {
		_ = d.DB.Close()
	}
	if d.AsynqClient != nil {
		_ = d.AsynqClient.Close()
	}
}
