package interfaces

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password, firstName, lastName, phone, address, city, state, zip, country string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, id, firstName, lastName, phone, address, city, state, zip, country string) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
}
