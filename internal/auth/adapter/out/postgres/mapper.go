package postgres

import (
	"database/sql"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
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

	var verificationCode *string
	if u.VerificationCode.Valid {
		s := u.VerificationCode.String
		verificationCode = &s
	}

	var verificationCodeExpiresAt *time.Time
	if u.VerificationCodeExpiresAt.Valid {
		t := u.VerificationCodeExpiresAt.Time
		verificationCodeExpiresAt = &t
	}

	return &domain.User{
		ID:                        u.ID,
		Email:                     u.Email,
		PasswordHash:              u.PasswordHash,
		IsMFAEnabled:              u.IsMfaEnabled,
		VerificationCode:          verificationCode,
		VerificationCodeExpiresAt: verificationCodeExpiresAt,
		Role:                      domain.UserRole(u.Role),
		Status:                    domain.UserStatus(u.Status),
		Profile:                   domain.Profile{Username: u.Username},
		CreatedAt:                 u.CreatedAt,
		UpdatedAt:                 u.UpdatedAt,
		LastLoginAt:               lastLoginAt,
		DeletedAt:                 deletedAt,
	}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
