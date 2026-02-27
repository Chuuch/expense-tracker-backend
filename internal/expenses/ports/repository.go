package ports

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
)

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, expense *domain.Expense) error
	GetExpenseByID(ctx context.Context, id string) (*domain.Expense, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Expense, error)
	DeleteExpense(ctx context.Context, id string) error
}