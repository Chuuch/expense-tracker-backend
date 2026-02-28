package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces/mocks"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"
)

func testConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			AppEnv:  "test",
			Version: "1.0.0",
		},
		Auth: config.AuthConfig{
			AccessTokenTTL: 15 * time.Minute,
		},
	}
}

func TestRegister_Success_Returns201(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	body := map[string]any{
		"email":      "test@example.com",
		"password":   "Password123!",
		"first_name": "Test",
		"last_name":  "User",
		"phone":      "+15550001111",
		"address":    "123 Main",
		"city":       "Austin",
		"state":      "TX",
		"zip":        "78701",
		"country":    "US",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Register(gomock.Any(), "test@example.com", "Password123!", "Test", "User", "+15550001111", "123 Main", "Austin", "TX", "78701", "US").
		Return(&domain.User{
			ID:           "u1",
			Email:        "test@example.com",
			Role:         domain.RoleUser,
			Status:       domain.StatusPending,
			IsMFAEnabled: false,
			Profile: domain.Profile{
				FirstName: "Test",
				LastName:  "User",
				Phone:     "+15550001111",
				Address:   "123 Main",
				City:      "Austin",
				State:     "TX",
				Zip:       "78701",
				Country:   "US",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil).
		Times(1)

	if err := h.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegister_Duplicate_Returns409(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	body := `{"email":"test@example.com","password":"Password123!","first_name":"A","last_name":"B","phone":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, usecase.ErrUserAlreadyExists).
		Times(1)

	if err := h.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegister_BindError_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("{invalid-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_InvalidCredentials_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	reqBody := `{"email":"test@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Login(gomock.Any(), "test@example.com", "wrong").
		Return(nil, usecase.ErrInvalidCredentials).
		Times(1)

	if err := h.Login(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_Success_Returns200AndToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	cfg := testConfig()
	h := authhttp.NewUserHandler(userUC, tokenUC, cfg)
	e := echo.New()

	reqBody := `{"email":"test@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user := &domain.User{
		ID:     "u1",
		Email:  "test@example.com",
		Role:   domain.RoleUser,
		Status: domain.UserStatus("pending"),
		Profile: domain.Profile{
			FirstName: "Test",
			LastName:  "User",
			Phone:     "+15550001111",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userUC.EXPECT().
		Login(gomock.Any(), "test@example.com", "Password123!").
		Return(user, nil).
		Times(1)

	tokenUC.EXPECT().
		GenerateToken(user, cfg.Auth.AccessTokenTTL).
		Return("mock-token", nil).
		Times(1)

	if err := h.Login(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetByID_NotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/missing-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/users/:id")
	c.SetPathValues(echo.PathValues{
		{Name: "id", Value: "missing-id"},
	})

	userUC.EXPECT().
		GetByID(gomock.Any(), "missing-id").
		Return(nil, usecase.ErrUserNotFound).
		Times(1)

	if err := h.GetByID(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUser_NotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/missing-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/users/:id")
	c.SetPathValues(echo.PathValues{
		{Name: "id", Value: "missing-id"},
	})

	userUC.EXPECT().
		DeleteUser(gomock.Any(), "missing-id").
		Return(usecase.ErrUserNotFound).
		Times(1)

	if err := h.DeleteUser(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// Optional sanity check that context gets passed through.
func TestRegister_UsesRequestContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := echo.New()

	reqBody := `{"email":"test@example.com","password":"Password123!","first_name":"T","last_name":"U","phone":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Register(gomock.AssignableToTypeOf(context.Background()), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&domain.User{
			ID:     "u1",
			Email:  "test@example.com",
			Role:   domain.RoleUser,
			Status: domain.StatusPending,
			Profile: domain.Profile{
				FirstName: "T",
				LastName:  "U",
				Phone:     "123",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil).
		Times(1)

	if err := h.Register(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}
