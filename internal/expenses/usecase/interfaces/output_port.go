package interfaces

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
)

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error)
	GetExpenseByID(ctx context.Context, userID, expenseID string) (*domain.Expense, error)
	ListExpenses(ctx context.Context, filter ExpenseListFilter) ([]*domain.Expense, error)
	UpdateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, userID, expenseID string) error
}
