package errors_test

import (
	"errors"
	"net/http"
	"testing"

	expenseerrors "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http/errors"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase"
)

func TestMap_NotFound_Returns404(t *testing.T) {
	code, payload := expenseerrors.Map(usecase.ErrExpenseNotFound)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
	if payload.Error != "Expense not found" {
		t.Fatalf("expected not found message, got %q", payload.Error)
	}
}

func TestMap_InvalidFilter_Returns400(t *testing.T) {
	code, payload := expenseerrors.Map(usecase.ErrExpenseInvalidFilter)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", code)
	}
	if payload.Error != "Invalid filters" {
		t.Fatalf("expected invalid filters message, got %q", payload.Error)
	}
}

func TestMap_Forbidden_Returns403(t *testing.T) {
	code, payload := expenseerrors.Map(usecase.ErrExpenseForbidden)
	if code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
	if payload.Error != "Forbidden" {
		t.Fatalf("expected forbidden message, got %q", payload.Error)
	}
}

func TestMap_WrappedDomainValidationError_Returns400(t *testing.T) {
	code, payload := expenseerrors.Map(errors.Join(errors.New("wrapped"), domain.ErrExpenseAmountInvalid))
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", code)
	}
	if payload.Error != "Invalid expense payload" {
		t.Fatalf("expected invalid payload message, got %q", payload.Error)
	}
}

func TestMap_Unknown_Returns500(t *testing.T) {
	code, payload := expenseerrors.Map(errors.New("boom"))
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", code)
	}
	if payload.Error != "Internal server error" {
		t.Fatalf("expected internal server error message, got %q", payload.Error)
	}
}
