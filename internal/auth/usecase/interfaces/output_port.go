package interfaces

import (
	"context"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string, replacedBy *string) error
	DeleteExpired(ctx context.Context) error
}

type TokenUsecase interface {
	GenerateToken(user *domain.User, duration time.Duration) (string, error)
	VerifyToken(token string) (*domain.TokenClaims, error)
}
