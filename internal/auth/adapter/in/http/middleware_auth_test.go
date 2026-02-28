package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces/mocks"
	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokenUC := mocks.NewMockTokenUsecase(ctrl)
	mw := authhttp.AuthMiddleware(tokenUC)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthMiddleware_InvalidScheme_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokenUC := mocks.NewMockTokenUsecase(ctrl)
	mw := authhttp.AuthMiddleware(tokenUC)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic abc")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokenUC := mocks.NewMockTokenUsecase(ctrl)
	mw := authhttp.AuthMiddleware(tokenUC)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tokenUC.EXPECT().
		VerifyToken("bad-token").
		Return("", errors.New("invalid token")).
		Times(1)

	next := func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthMiddleware_ValidToken_CallsNextAndSetsUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tokenUC := mocks.NewMockTokenUsecase(ctrl)
	mw := authhttp.AuthMiddleware(tokenUC)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tokenUC.EXPECT().
		VerifyToken("good-token").
		Return("user-123", nil).
		Times(1)

	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true

		// Check context value set by middleware
		if got, ok := authhttp.GetUserIDFromContext(c.Request().Context()); !ok || got != "user-123" {
			t.Fatalf("expected user_id in context, got=%q ok=%v", got, ok)
		}
		return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
	}

	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["ok"] != "true" {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}
