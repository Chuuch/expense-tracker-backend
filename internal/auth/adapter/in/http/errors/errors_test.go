package http_test

import (
	stderrors "errors"
	"net/http"
	"testing"

	httperrors "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http/errors"
	authusecase "github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
)

func TestMap_UserAlreadyExists(t *testing.T) {
	code, resp := httperrors.Map(authusecase.ErrUserAlreadyExists)
	if code != http.StatusConflict || resp.Error != "User already exists" {
		t.Fatalf("unexpected map result: code=%d resp=%+v", code, resp)
	}
}

func TestMap_InvalidCredentials(t *testing.T) {
	code, resp := httperrors.Map(authusecase.ErrInvalidCredentials)
	if code != http.StatusUnauthorized || resp.Error != "Invalid credentials" {
		t.Fatalf("unexpected map result: code=%d resp=%+v", code, resp)
	}
}

func TestMap_UserNotFound(t *testing.T) {
	code, resp := httperrors.Map(authusecase.ErrUserNotFound)
	if code != http.StatusNotFound || resp.Error != "User not found" {
		t.Fatalf("unexpected map result: code=%d resp=%+v", code, resp)
	}
}

func TestMap_UnknownError(t *testing.T) {
	code, resp := httperrors.Map(stderrors.New("boom"))
	if code != http.StatusInternalServerError || resp.Error != "Internal server error" {
		t.Fatalf("unexpected map result: code=%d resp=%+v", code, resp)
	}
}
