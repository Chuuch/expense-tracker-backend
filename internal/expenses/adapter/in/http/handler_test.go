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
	expensehttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces/mocks"
	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"
)

func newExpenseEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewEchoValidator()
	return e
}

func withAuthUser(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), authhttp.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestCreateExpense_Success_Returns201(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	body := map[string]any{
		"amount":      12345,
		"currency":    "usd",
		"category":    "housing",
		"description": "Rent",
		"date":        "2026-03-01T10:00:00Z",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/expenses", bytes.NewReader(b))
	req = withAuthUser(req, "u1")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expenseUC.EXPECT().
		CreateExpense(
			gomock.Any(),
			"u1",
			int64(12345),
			"usd",
			domain.CategoryHousing,
			"Rent",
			time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		).
		Return(&domain.Expense{
			ID:          "e1",
			UserID:      "u1",
			Amount:      12345,
			Currency:    "USD",
			Category:    domain.CategoryHousing,
			Description: "Rent",
			Date:        time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil).
		Times(1)

	if err := h.CreateExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateExpense_Unauthorized_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/expenses", bytes.NewBufferString(`{
		"amount":12345,
		"currency":"usd",
		"category":"housing",
		"description":"Rent",
		"date":"2026-03-01T10:00:00Z"
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expenseUC.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	if err := h.CreateExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListExpenses_WithCategoryFilter_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/expenses?category=housing&limit=10&offset=0",
		nil,
	)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expenseUC.EXPECT().
		ListExpenses(gomock.Any(), gomock.AssignableToTypeOf(interfaces.ExpenseListFilter{})).
		DoAndReturn(func(_ context.Context, f interfaces.ExpenseListFilter) ([]*domain.Expense, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if f.Category == nil || *f.Category != domain.CategoryHousing {
				t.Fatalf("expected category housing, got %+v", f.Category)
			}
			if f.Limit != 10 || f.Offset != 0 {
				t.Fatalf("expected limit/offset 10/0, got %d/%d", f.Limit, f.Offset)
			}
			return []*domain.Expense{
				{
					ID:          "e1",
					UserID:      "u1",
					Amount:      12345,
					Currency:    "USD",
					Category:    domain.CategoryHousing,
					Description: "Rent",
					Date:        time.Now(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			}, nil
		}).
		Times(1)

	if err := h.ListExpenses(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetExpenseByID_NotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/expenses/e_missing", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/expenses/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "e_missing"}})

	expenseUC.EXPECT().
		GetExpenseByID(gomock.Any(), "u1", "e_missing").
		Return(nil, usecase.ErrExpenseNotFound).
		Times(1)

	if err := h.GetExpenseByID(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateExpense_InvalidCategory_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	body := `{"amount":12345,"currency":"usd","category":"invalid-cat","description":"x","date":"2026-03-01T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/expenses", bytes.NewBufferString(body))
	req = withAuthUser(req, "u1")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expenseUC.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	if err := h.CreateExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateExpense_Success_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	body := `{"amount":2000,"currency":"usd","category":"housing","description":"Rent","date":"2026-03-01T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/expenses/e1", bytes.NewBufferString(body))
	req = withAuthUser(req, "u1")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/expenses/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "e1"}})

	expenseUC.EXPECT().
		UpdateExpense(
			gomock.Any(),
			"u1",
			"e1",
			int64(2000),
			"usd",
			domain.CategoryHousing,
			"Rent",
			time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		).
		Return(&domain.Expense{
			ID:          "e1",
			UserID:      "u1",
			Amount:      2000,
			Currency:    "USD",
			Category:    domain.CategoryHousing,
			Description: "Rent",
			Date:        time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil).
		Times(1)

	if err := h.UpdateExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateExpense_NotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	body := `{"amount":2000,"currency":"usd","category":"housing","description":"Rent","date":"2026-03-01T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/expenses/e_missing", bytes.NewBufferString(body))
	req = withAuthUser(req, "u1")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/expenses/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "e_missing"}})

	expenseUC.EXPECT().
		UpdateExpense(gomock.Any(), "u1", "e_missing", int64(2000), "usd", domain.CategoryHousing, "Rent", time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)).
		Return(nil, usecase.ErrExpenseNotFound).
		Times(1)

	if err := h.UpdateExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeleteExpense_Success_Returns204(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/expenses/e1", nil)
	req = withAuthUser(req, "u1")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/expenses/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "e1"}})

	expenseUC.EXPECT().
		DeleteExpense(gomock.Any(), "u1", "e1").
		Return(nil).
		Times(1)

	if err := h.DeleteExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDeleteExpense_NotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expenseUC := mocks.NewMockExpenseUsecase(ctrl)
	h := expensehttp.NewExpenseHandler(expenseUC)
	e := newExpenseEcho()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/expenses/e_missing", nil)
	req = withAuthUser(req, "u1")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/expenses/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "e_missing"}})

	expenseUC.EXPECT().
		DeleteExpense(gomock.Any(), "u1", "e_missing").
		Return(usecase.ErrExpenseNotFound).
		Times(1)

	if err := h.DeleteExpense(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}
