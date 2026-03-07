package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	q postgresdb.Querier
}

func NewUserRepository(q postgresdb.Querier) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, postgresdb.CreateUserParams{
		ID:                        user.ID,
		Email:                     user.Email,
		PasswordHash:              user.PasswordHash,
		IsMfaEnabled:              user.IsMFAEnabled,
		VerificationCode:          toNullString(user.VerificationCode),
		VerificationCodeExpiresAt: toNullTime(user.VerificationCodeExpiresAt),
		Role:                      string(user.Role),
		Status:                    string(user.Status),
		Username:                  user.Profile.Username,
		CreatedAt:                 user.CreatedAt,
		UpdatedAt:                 user.UpdatedAt,
		LastLoginAt:               toNullTime(user.LastLoginAt),
		DeletedAt:                 toNullTime(user.DeletedAt),
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return r.q.UpdateUser(ctx, postgresdb.UpdateUserParams{
		ID:                        user.ID,
		Email:                     user.Email,
		PasswordHash:              user.PasswordHash,
		IsMfaEnabled:              user.IsMFAEnabled,
		VerificationCode:          toNullString(user.VerificationCode),
		VerificationCodeExpiresAt: toNullTime(user.VerificationCodeExpiresAt),
		Role:                      string(user.Role),
		Status:                    string(user.Status),
		Username:                  user.Profile.Username,
		UpdatedAt:                 user.UpdatedAt,
		LastLoginAt:               toNullTime(user.LastLoginAt),
	})
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	err := r.q.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
