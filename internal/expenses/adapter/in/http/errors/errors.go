package errors

import (
	"errors"
	"net/http"

	expensedomain "github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	expenseusecase "github.com/chuuch/expense-tracker-backend/internal/expenses/usecase"
)

type Response struct {
	Error string `json:"error"`
}

func Map(err error) (int, Response) {
	switch {
	case errors.Is(err, expenseusecase.ErrExpenseNotFound):
		return http.StatusNotFound, Response{Error: "Expense not found"}

	case errors.Is(err, expenseusecase.ErrExpenseForbidden):
		return http.StatusForbidden, Response{Error: "Forbidden"}

	case errors.Is(err, expenseusecase.ErrExpenseInvalidFilter):
		return http.StatusBadRequest, Response{Error: "Invalid filters"}

	case errors.Is(err, expensedomain.ErrExpenseIDRequired),
		errors.Is(err, expensedomain.ErrExpenseUserIDRequired),
		errors.Is(err, expensedomain.ErrExpenseAmountInvalid),
		errors.Is(err, expensedomain.ErrExpenseCurrencyEmpty),
		errors.Is(err, expensedomain.ErrExpenseDateRequired):
		return http.StatusBadRequest, Response{Error: "Invalid expense payload"}

	default:
		return http.StatusInternalServerError, Response{Error: "Internal server error"}
	}
}
