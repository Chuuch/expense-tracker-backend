package ports

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, eail string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
}
