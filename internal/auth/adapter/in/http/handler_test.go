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
	"github.com/chuuch/expense-tracker-backend/utils"
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
	e := newEchoWithValidator()

	body := map[string]any{
		"email":    "test@example.com",
		"password": "Password123!",
		"username": "Test",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Register(gomock.Any(), "test@example.com", "Password123!", "Test").
		Return(&domain.User{
			ID:           "u1",
			Email:        "test@example.com",
			Role:         domain.RoleUser,
			Status:       domain.StatusPending,
			IsMFAEnabled: false,
			Profile: domain.Profile{
				Username: "Test",
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
	e := newEchoWithValidator()

	body := `{"email":"test@example.com","password":"Password123!","username":"A"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
	e := newEchoWithValidator()

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
	e := newEchoWithValidator()

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
	e := newEchoWithValidator()

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
			Username: "Test",
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
	e := newEchoWithValidator()

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
	e := newEchoWithValidator()

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
	e := newEchoWithValidator()

	reqBody := `{"email":"test@example.com","password":"Password123!","username":"T"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Register(gomock.AssignableToTypeOf(context.Background()), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&domain.User{
			ID:     "u1",
			Email:  "test@example.com",
			Role:   domain.RoleUser,
			Status: domain.StatusPending,
			Profile: domain.Profile{
				Username: "T",
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

type refreshUsecaseStub struct {
	issueTokenPairFn func(ctx context.Context, user *domain.User) (*domain.TokenPair, error)
	refreshFn        func(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error)
	revokeFn         func(ctx context.Context, rawRefreshToken string) error
	deleteExpiredFn  func(ctx context.Context) error
}

type accessBlacklistStub struct {
	addCalled bool
	addToken  string
	addTTL    time.Duration
	addErr    error
}

func (s *accessBlacklistStub) Add(_ context.Context, token string, ttl time.Duration) error {
	s.addCalled = true
	s.addToken = token
	s.addTTL = ttl
	return s.addErr
}

func (s *accessBlacklistStub) Contains(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (s *refreshUsecaseStub) IssueTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	if s.issueTokenPairFn != nil {
		return s.issueTokenPairFn(ctx, user)
	}
	return nil, nil
}

func (s *refreshUsecaseStub) Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error) {
	if s.refreshFn != nil {
		return s.refreshFn(ctx, rawRefreshToken)
	}
	return nil, nil
}

func (s *refreshUsecaseStub) Revoke(ctx context.Context, rawRefreshToken string) error {
	if s.revokeFn != nil {
		return s.revokeFn(ctx, rawRefreshToken)
	}
	return nil
}

func (s *refreshUsecaseStub) DeleteExpired(ctx context.Context) error {
	if s.deleteExpiredFn != nil {
		return s.deleteExpiredFn(ctx)
	}
	return nil
}

func newEchoWithValidator() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewEchoValidator()
	return e
}

func TestRefresh_NotConfigured_Returns501(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig()) // no refresh usecase wired
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Refresh(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status 501, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRefresh_InvalidToken_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	refreshUC := &refreshUsecaseStub{
		refreshFn: func(ctx context.Context, raw string) (*domain.TokenPair, error) {
			if raw != "rt-old" {
				t.Fatalf("expected refresh token rt-old, got %q", raw)
			}
			return nil, usecase.ErrRefreshTokenInvalid
		},
	}

	h := authhttp.NewUserHandlerWithRefresh(userUC, tokenUC, refreshUC, testConfig())
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Refresh(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRefresh_Success_Returns200AndPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	refreshUC := &refreshUsecaseStub{
		refreshFn: func(ctx context.Context, raw string) (*domain.TokenPair, error) {
			if raw != "rt-old" {
				t.Fatalf("expected refresh token rt-old, got %q", raw)
			}
			return &domain.TokenPair{
				AccessToken:  "at-new",
				RefreshToken: "rt-new",
			}, nil
		},
	}

	h := authhttp.NewUserHandlerWithRefresh(userUC, tokenUC, refreshUC, testConfig())
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Refresh(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp authhttp.RefreshResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.AccessToken != "at-new" || resp.RefreshToken != "rt-new" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLogout_NotConfigured_Returns501(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig()) // no refresh usecase wired
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status 501, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogout_InvalidToken_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	refreshUC := &refreshUsecaseStub{
		revokeFn: func(ctx context.Context, raw string) error {
			if raw != "rt-old" {
				t.Fatalf("expected refresh token rt-old, got %q", raw)
			}
			return usecase.ErrRefreshTokenInvalid
		},
	}

	h := authhttp.NewUserHandlerWithRefresh(userUC, tokenUC, refreshUC, testConfig())
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogout_Success_Returns204(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	called := false
	refreshUC := &refreshUsecaseStub{
		revokeFn: func(ctx context.Context, raw string) error {
			called = true
			if raw != "rt-old" {
				t.Fatalf("expected refresh token rt-old, got %q", raw)
			}
			return nil
		},
	}

	h := authhttp.NewUserHandlerWithRefresh(userUC, tokenUC, refreshUC, testConfig())
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !called {
		t.Fatal("expected revoke to be called")
	}
}

func TestLogout_Success_BlacklistsAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	refreshUC := &refreshUsecaseStub{
		revokeFn: func(ctx context.Context, raw string) error {
			if raw != "rt-old" {
				t.Fatalf("expected refresh token rt-old, got %q", raw)
			}
			return nil
		},
	}

	cfg := testConfig()
	blacklist := &accessBlacklistStub{}
	h := authhttp.NewUserHandlerWithRefresh(userUC, tokenUC, refreshUC, cfg, blacklist)
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(`{"refresh_token":"rt-old"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer access-token-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !blacklist.addCalled {
		t.Fatalf("expected access token to be blacklisted")
	}
	if blacklist.addToken != "access-token-123" {
		t.Fatalf("expected blacklisted token access-token-123, got %s", blacklist.addToken)
	}
	if blacklist.addTTL != cfg.Auth.AccessTokenTTL {
		t.Fatalf("expected ttl %v, got %v", cfg.Auth.AccessTokenTTL, blacklist.addTTL)
	}
}

func TestVerifyEmail_Success_Returns200AndUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"test@example.com","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user := &domain.User{
		ID:        "u1",
		Email:     "test@example.com",
		Role:      domain.RoleUser,
		Status:    domain.StatusActive,
		Profile:   domain.Profile{Username: "Test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	userUC.EXPECT().
		VerifyEmail(gomock.Any(), "test@example.com", "123456").
		Return(user, nil).
		Times(1)

	if err := h.VerifyEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp authhttp.UserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID != "u1" || resp.Email != "test@example.com" || resp.Status != "active" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestVerifyEmail_InvalidCode_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"test@example.com","code":"999999"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		VerifyEmail(gomock.Any(), "test@example.com", "999999").
		Return(nil, usecase.ErrInvalidVerificationCode).
		Times(1)

	if err := h.VerifyEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestVerifyEmail_ExpiredCode_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"test@example.com","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		VerifyEmail(gomock.Any(), "test@example.com", "123456").
		Return(nil, usecase.ErrVerificationCodeExpired).
		Times(1)

	if err := h.VerifyEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestVerifyEmail_UserNotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"missing@example.com","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		VerifyEmail(gomock.Any(), "missing@example.com", "123456").
		Return(nil, usecase.ErrUserNotFound).
		Times(1)

	if err := h.VerifyEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestVerifyEmail_BindError_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBufferString("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.VerifyEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLogin_UserNotActive_Returns403(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"test@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		Login(gomock.Any(), "test@example.com", "Password123!").
		Return(nil, usecase.ErrUserNotActive).
		Times(1)

	if err := h.Login(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp authhttp.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != "Please verify your email to sign in" {
		t.Fatalf("unexpected error message: %q", resp.Error)
	}
}

func TestResendVerificationEmail_Success_Returns204(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		ResendVerificationEmail(gomock.Any(), "test@example.com").
		Return(nil).
		Times(1)

	if err := h.ResendVerificationEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestResendVerificationEmail_UserNotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"missing@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		ResendVerificationEmail(gomock.Any(), "missing@example.com").
		Return(usecase.ErrUserNotFound).
		Times(1)

	if err := h.ResendVerificationEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestResendVerificationEmail_AlreadyActive_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification-email", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userUC.EXPECT().
		ResendVerificationEmail(gomock.Any(), "test@example.com").
		Return(usecase.ErrUserAlreadyActive).
		Times(1)

	if err := h.ResendVerificationEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestResendVerificationEmail_BindError_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mocks.NewMockUserUsecase(ctrl)
	tokenUC := mocks.NewMockTokenUsecase(ctrl)

	h := authhttp.NewUserHandler(userUC, tokenUC, testConfig())
	e := newEchoWithValidator()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification-email", bytes.NewBufferString("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ResendVerificationEmail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}
