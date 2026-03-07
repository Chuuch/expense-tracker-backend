package interfaces

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password, username, phone, address, city, state, zip, country string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, id, username, phone, address, city, state, zip, country string) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	VerifyEmail(ctx context.Context, email, code string) (*domain.User, error)
}

type RefreshTokenUsecase interface {
	IssueTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error)
	Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error)
	Revoke(ctx context.Context, rawRefreshToken string) error
	DeleteExpired(ctx context.Context) error
}
