package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	authdomain "github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	authinterfaces "github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	expensehttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	expensedomain "github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	expenseinterfaces "github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

func newRouterSmokeApp() *App {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         ":0",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
		Auth: config.AuthConfig{
			RateLimitRequests: 10,
			RateLimitWindow:   time.Minute,
		},
	}

	a := NewApp(
		cfg,
		zap.NewNop(),
		&authhttp.UserHandler{},
		&expensehttp.ExpenseHandler{},
		nil, // goalHandler
		nil, // analyticsHandler
		nil, // tokenUsecase
		nil, // accessTokenBlacklist
		nil, // db
		nil, // redis
	)
	a.registerRoutes()
	return a
}

type tokenUsecaseStub struct {
	verifyFn func(token string) (*authdomain.TokenClaims, error)
}

func (s tokenUsecaseStub) GenerateToken(_ *authdomain.User, _ time.Duration) (string, error) {
	return "", errors.New("not implemented")
}

func (s tokenUsecaseStub) VerifyToken(token string) (*authdomain.TokenClaims, error) {
	if s.verifyFn == nil {
		return nil, errors.New("verify function not set")
	}
	return s.verifyFn(token)
}

type expenseUsecaseStub struct{}

func (expenseUsecaseStub) CreateExpense(
	_ context.Context,
	_ string,
	_ int64,
	_ string,
	_ expensedomain.ExpenseCategory,
	_ string,
	_ time.Time,
) (*expensedomain.Expense, error) {
	return nil, errors.New("not implemented")
}

func (expenseUsecaseStub) GetExpenseByID(_ context.Context, _, _ string) (*expensedomain.Expense, error) {
	return nil, errors.New("not implemented")
}

func (expenseUsecaseStub) ListExpenses(
	_ context.Context,
	_ expenseinterfaces.ExpenseListFilter,
) ([]*expensedomain.Expense, error) {
	return []*expensedomain.Expense{}, nil
}

func (expenseUsecaseStub) UpdateExpense(
	_ context.Context,
	_ string,
	_ string,
	_ int64,
	_ string,
	_ expensedomain.ExpenseCategory,
	_ string,
	_ time.Time,
) (*expensedomain.Expense, error) {
	return nil, errors.New("not implemented")
}

func (expenseUsecaseStub) DeleteExpense(_ context.Context, _, _ string) error {
	return errors.New("not implemented")
}

var (
	_ authinterfaces.TokenUsecase      = tokenUsecaseStub{}
	_ expenseinterfaces.ExpenseUsecase = expenseUsecaseStub{}
)

func TestRouterSmoke_Health_Returns200(t *testing.T) {
	a := newRouterSmokeApp()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	a.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterSmoke_HealthLive_Returns200(t *testing.T) {
	a := newRouterSmokeApp()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	a.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterSmoke_HealthReady_WithoutDeps_Returns503(t *testing.T) {
	a := newRouterSmokeApp()

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	a.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterSmoke_ExpensesRoutes_RequireAuth(t *testing.T) {
	a := newRouterSmokeApp()

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create", method: http.MethodPost, path: "/api/v1/expenses"},
		{name: "list", method: http.MethodGet, path: "/api/v1/expenses"},
		{name: "getByID", method: http.MethodGet, path: "/api/v1/expenses/e1"},
		{name: "update", method: http.MethodPut, path: "/api/v1/expenses/e1"},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/expenses/e1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			a.echo.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRouterSmoke_ExpensesList_WithValidBearer_NotUnauthorized(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         ":0",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
		Auth: config.AuthConfig{
			RateLimitRequests: 10,
			RateLimitWindow:   time.Minute,
		},
	}

	tokenUC := tokenUsecaseStub{
		verifyFn: func(token string) (*authdomain.TokenClaims, error) {
			if token != "valid-token" {
				return nil, errors.New("invalid token")
			}
			return &authdomain.TokenClaims{
				UserID: "u1",
				Role:   authdomain.RoleUser,
			}, nil
		},
	}

	expenseHandler := expensehttp.NewExpenseHandler(expenseUsecaseStub{})

	a := NewApp(
		cfg,
		zap.NewNop(),
		&authhttp.UserHandler{},
		expenseHandler,
		nil, // goalHandler
		nil, // analyticsHandler
		tokenUC,
		nil, // accessTokenBlacklist
		nil, // db
		nil, // redis
	)
	a.registerRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expenses", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	a.echo.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("expected non-401 response, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterSmoke_AuthRateLimiter_SecondRequestReturns429(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         ":0",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
		Auth: config.AuthConfig{
			RateLimitRequests: 1,
			RateLimitWindow:   time.Minute,
		},
	}

	a := NewApp(
		cfg,
		zap.NewNop(),
		&authhttp.UserHandler{},
		&expensehttp.ExpenseHandler{},
		nil, // goalHandler
		nil, // analyticsHandler
		nil, // tokenUsecase
		nil, // accessTokenBlacklist
		nil, // db
		nil, // redis
	)
	a.registerRoutes()

	body := "{invalid-json"

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	a.echo.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusBadRequest {
		t.Fatalf("expected first request status 400, got %d body=%s", rec1.Code, rec1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	a.echo.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestRouterSmoke_PanicRecovered_ReturnsNormalized500(t *testing.T) {
	a := newRouterSmokeApp()
	a.echo.Use(middleware.Recover())
	a.echo.Use(authhttp.RequestIDMiddleware())
	a.echo.GET("/panic", func(c *echo.Context) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	a.echo.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Internal server error") {
		t.Fatalf("expected normalized internal error payload, got body=%s", rec.Body.String())
	}
}

func TestRouterSmoke_GoalsRoutes_RequireAuth(t *testing.T) {
	a := newRouterSmokeApp()

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create", method: http.MethodPost, path: "/api/v1/goals"},
		{name: "list", method: http.MethodGet, path: "/api/v1/goals"},
		{name: "getByID", method: http.MethodGet, path: "/api/v1/goals/g1"},
		{name: "update", method: http.MethodPatch, path: "/api/v1/goals/g1"},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/goals/g1"},
		{name: "addContribution", method: http.MethodPost, path: "/api/v1/goals/g1/contributions"},
		{name: "listContributions", method: http.MethodGet, path: "/api/v1/goals/g1/contributions"},
		{name: "progress", method: http.MethodGet, path: "/api/v1/goals/g1/progress"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			a.echo.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRouterSmoke_AuthGoogleRoutes_Registered(t *testing.T) {
	a := newRouterSmokeApp()

	for _, path := range []string{
		"/api/v1/auth/google",
		"/api/v1/auth/google/callback",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		a.echo.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			t.Fatalf("expected route %s to be registered, got 404 body=%s", path, rec.Body.String())
		}
	}
}
