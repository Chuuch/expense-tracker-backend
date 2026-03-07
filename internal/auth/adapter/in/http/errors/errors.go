package http

import (
	"errors"
	"net/http"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	authusecase "github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
)

type Response struct {
	Error string `json:"error"`
}

func Map(err error) (int, Response) {
	switch {
	case errors.Is(err, authusecase.ErrUserAlreadyExists), errors.Is(err, domain.ErrUserAlreadyExists):
		return http.StatusConflict, Response{Error: "User already exists"}

	case errors.Is(err, authusecase.ErrInvalidCredentials), errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, Response{Error: "Invalid credentials"}

	case errors.Is(err, authusecase.ErrUserNotFound), errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, Response{Error: "User not found"}

	case errors.Is(err, authusecase.ErrRefreshTokenInvalid), errors.Is(err, domain.ErrRefreshTokenInvalid):
		return http.StatusUnauthorized, Response{Error: "Invalid refresh token"}

	case errors.Is(err, authusecase.ErrRefreshTokenExpired), errors.Is(err, domain.ErrRefreshTokenExpired):
		return http.StatusUnauthorized, Response{Error: "Refresh token expired"}

	case errors.Is(err, authusecase.ErrRefreshTokenRevoked), errors.Is(err, domain.ErrRefreshTokenRevoked):
		return http.StatusUnauthorized, Response{Error: "Refresh token revoked"}
	case errors.Is(err, authusecase.ErrInvalidVerificationCode), errors.Is(err, domain.ErrInvalidVerificationCode):
		return http.StatusBadRequest, Response{Error: "Invalid verification code"}

	case errors.Is(err, authusecase.ErrVerificationCodeExpired), errors.Is(err, domain.ErrVerificationCodeExpired):
		return http.StatusBadRequest, Response{Error: "Verification code expired"}

	default:
		return http.StatusInternalServerError, Response{Error: "Internal server error"}
	}
}
