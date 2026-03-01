package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	authdomain "github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	authinterfaces "github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	expensehttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	expensedomain "github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	expenseinterfaces "github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
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
	}

	a := NewApp(
		cfg,
		zap.NewNop(),
		&authhttp.UserHandler{},
		&expensehttp.ExpenseHandler{},
		nil,
		nil,
		nil,
		nil,
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
		tokenUC,
		nil,
		nil,
		nil,
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
