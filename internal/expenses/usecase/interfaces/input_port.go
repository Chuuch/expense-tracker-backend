package interfaces

import (
	"context"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
)

type ExpenseListFilter struct {
	UserID string
	Category *domain.ExpenseCategory
	FromDate *time.Time
	ToDate *time.Time
	Limit int
	Offset int
}

type ExpenseUsecase interface {
	CreateExpense(
		ctx context.Context,
		userID string,
		amount int64,
		currency string,
		category domain.ExpenseCategory,
		description string,
		date time.Time,
	) (*domain.Expense, error)

	GetExpenseByID(ctx context.Context, userID, expenseID string) (*domain.Expense, error)
	ListExpenses(ctx context.Context, filter ExpenseListFilter) ([]*domain.Expense, error)

	UpdateExpense(
		ctx context.Context,
		userID string,
		expenseID string,
		amount int64,
		currency string,
		category domain.ExpenseCategory,
		description string,
		date time.Time,
	) (*domain.Expense, error)

	DeleteExpense(ctx context.Context, userID, expenseID string) error
}