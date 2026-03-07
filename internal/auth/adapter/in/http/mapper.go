package http

import (
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

func mapUserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		Username:    u.Profile.Username,
		Role:         string(u.Role),
		Status:       string(u.Status),
		IsMFAEnabled: u.IsMFAEnabled,
		CreatedAt:    u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
