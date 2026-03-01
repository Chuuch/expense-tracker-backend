package http

import (
	"fmt"
	"strings"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
)

func mapExpenseToResponse(e *domain.Expense) ExpenseResponse {
	return ExpenseResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		Amount:      e.Amount,
		Currency:    e.Currency,
		Category:    string(e.Category),
		Description: e.Description,
		Date:        e.Date.UTC().Format(time.RFC3339),
		CreatedAt:   e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapExpnsesToResponse(expenses []*domain.Expense, limit, offset int) ListExpensesResponse {
	responses := make([]ExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		responses = append(responses, mapExpenseToResponse(e))
	}

	return ListExpensesResponse{
		Expenses: responses,
		Limit:    limit,
		Offset:   offset,
		Count:    len(expenses),
	}
}

func parseCategory(raw string) (domain.ExpenseCategory, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return "", nil
	}

	c := domain.ExpenseCategory(v)
	switch c {
	case domain.CategoryFood,
		domain.CategoryTransport,
		domain.CategoryHousing,
		domain.CategoryUtilities,
		domain.CategoryHealthcare,
		domain.CategoryEntertainment,
		domain.CategoryOther:
		return c, nil
	default:
		return "", fmt.Errorf("invalid expense category: %s", raw)
	}
}
