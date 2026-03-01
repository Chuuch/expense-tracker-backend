package postgres

import (
	"database/sql"
	"time"

	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

func toDomainUser(u postgresdb.User) *domain.User {
	var lastLoginAt *time.Time
	if u.LastLoginAt.Valid {
		t := u.LastLoginAt.Time
		lastLoginAt = &t
	}

	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		deletedAt = &t
	}

	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsMFAEnabled: u.IsMfaEnabled,
		Role:         domain.UserRole(u.Role),
		Status:       domain.UserStatus(u.Status),
		Profile: domain.Profile{
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Phone:     u.Phone,
			Address:   u.Address,
			City:      u.City,
			State:     u.State,
			Zip:       u.Zip,
			Country:   u.Country,
		},
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: lastLoginAt,
		DeletedAt:   deletedAt,
	}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
