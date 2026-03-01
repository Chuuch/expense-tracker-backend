package app

import (
	"database/sql"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	expensehttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	DB                   *sql.DB
	Redis                *redis.Client
	UserHandler          *authhttp.UserHandler
	ExpenseHandler       *expensehttp.ExpenseHandler
	TokenUsecase         interfaces.TokenUsecase
	AccessTokenBlacklist interfaces.AccessTokenBlacklist
}

func (d *Dependencies) Close() {
	if d.Redis != nil {
		_ = d.Redis.Close()
	}
	if d.DB != nil {
		_ = d.DB.Close()
	}
}
