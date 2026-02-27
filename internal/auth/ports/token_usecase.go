package ports

import (
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

type TokenUsecase interface {
	GenerateToken(user *domain.User, duration time.Duration) (string, error)
	VerifyToken(token string) (string, error)
}
