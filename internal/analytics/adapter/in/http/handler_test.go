package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	analyticshttp "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces/mocks"
	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"
)

func newAnalyticsEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewEchoValidator()
	return e
}

func withAuthUser(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), authhttp.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestGetMonthlySpending_Unauthorized_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAnalyticsUsecase(ctrl)
	h := analyticshttp.NewAnalyticsHandler(uc)
	e := newAnalyticsEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/monthly?from_date=2025-01-01T00:00:00Z&to_date=2025-12-31T23:59:59Z", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	uc.EXPECT().GetMonthlySpending(gomock.Any(), gomock.Any()).Times(0)

	if err := h.GetMonthlySpending(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetMonthlySpending_InvalidQueryParams_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAnalyticsUsecase(ctrl)
	h := analyticshttp.NewAnalyticsHandler(uc)
	e := newAnalyticsEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/monthly?from_date=not-a-date&to_date=2025-12-31T23:59:59Z", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	uc.EXPECT().GetMonthlySpending(gomock.Any(), gomock.Any()).Times(0)

	if err := h.GetMonthlySpending(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetMonthlySpending_Success_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAnalyticsUsecase(ctrl)
	h := analyticshttp.NewAnalyticsHandler(uc)
	e := newAnalyticsEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/monthly?from_date=2025-01-01T00:00:00Z&to_date=2025-06-30T23:59:59Z", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)

	uc.EXPECT().
		GetMonthlySpending(
			gomock.Any(),
			gomock.AssignableToTypeOf(interfaces.MonthlySpendingFilter{}),
		).
		DoAndReturn(func(_ context.Context, f interfaces.MonthlySpendingFilter) ([]interfaces.MonthlySpendingPoint, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if !f.From.Equal(from) || !f.To.Equal(to) {
				t.Fatalf("expected from/to 2025-01-01 and 2025-06-30, got %v / %v", f.From, f.To)
			}
			return []interfaces.MonthlySpendingPoint{
				{YearMonth: "2025-01", Amount: 1500, Currency: "USD"},
				{YearMonth: "2025-02", Amount: 2000, Currency: "USD"},
			}, nil
		}).
		Times(1)

	if err := h.GetMonthlySpending(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() == "" {
		t.Fatal("expected non-empty JSON response")
	}
	if !contains(rec.Body.String(), "2025-01") || !contains(rec.Body.String(), "1500") {
		t.Fatalf("expected response to contain year_month and amount, got %s", rec.Body.String())
	}
}

func TestGetSpendingByCategory_Unauthorized_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAnalyticsUsecase(ctrl)
	h := analyticshttp.NewAnalyticsHandler(uc)
	e := newAnalyticsEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/by-category?from_date=2025-01-01T00:00:00Z&to_date=2025-12-31T23:59:59Z", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	uc.EXPECT().GetSpendingByCategory(gomock.Any(), gomock.Any()).Times(0)

	if err := h.GetSpendingByCategory(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetSpendingByCategory_InvalidQueryParams_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAnalyticsUsecase(ctrl)
	h := analyticshttp.NewAnalyticsHandler(uc)
	e := newAnalyticsEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/by-category?from_date=2025-01-01T00:00:00Z&to_date=invalid", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	uc.EXPECT().GetSpendingByCategory(gomock.Any(), gomock.Any()).Times(0)

	if err := h.GetSpendingByCategory(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetSpendingByCategory_Success_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := mocks.NewMockAnalyticsUsecase(ctrl)
	h := analyticshttp.NewAnalyticsHandler(uc)
	e := newAnalyticsEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/by-category?from_date=2025-03-01T00:00:00Z&to_date=2025-03-31T23:59:59Z", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	from := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)

	uc.EXPECT().
		GetSpendingByCategory(
			gomock.Any(),
			gomock.AssignableToTypeOf(interfaces.CategorySpendingFilter{}),
		).
		DoAndReturn(func(_ context.Context, f interfaces.CategorySpendingFilter) ([]interfaces.CategorySpendingPoint, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if !f.FromDate.Equal(from) || !f.ToDate.Equal(to) {
				t.Fatalf("expected from/to 2025-03-01 and 2025-03-31, got %v / %v", f.FromDate, f.ToDate)
			}
			return []interfaces.CategorySpendingPoint{
				{Category: "housing", Amount: 3000, Currency: "USD"},
				{Category: "food", Amount: 1200, Currency: "USD"},
			}, nil
		}).
		Times(1)

	if err := h.GetSpendingByCategory(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !contains(rec.Body.String(), "housing") || !contains(rec.Body.String(), "3000") {
		t.Fatalf("expected response to contain category and amount, got %s", rec.Body.String())
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
