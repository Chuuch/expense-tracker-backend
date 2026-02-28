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
	default:
		return http.StatusInternalServerError, Response{Error: "Internal server error"}
	}
}
