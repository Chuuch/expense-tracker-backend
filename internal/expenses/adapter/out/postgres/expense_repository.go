package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

type ExpenseRepository struct {
	q postgresdb.Querier
}

func NewExpenseRepository(q postgresdb.Querier) *ExpenseRepository {
	return &ExpenseRepository{q: q}
}

var _ interfaces.ExpenseRepository = (*ExpenseRepository)(nil)

func (r *ExpenseRepository) CreateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error) {
	row, err := r.q.CreateExpense(ctx, postgresdb.CreateExpenseParams{
		ID:          expense.ID,
		UserID:      expense.UserID,
		Amount:      expense.Amount,
		Currency:    expense.Currency,
		Category:    string(expense.Category),
		Description: expense.Description,
		Date:        expense.Date,
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}
	return toDomainExpense(row), nil
}

func (r *ExpenseRepository) GetExpenseByID(ctx context.Context, userID, expenseID string) (*domain.Expense, error) {
	row, err := r.q.GetExpenseByID(ctx, postgresdb.GetExpenseByIDParams{
		ID:     expenseID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get expense by id: %w", err)
	}
	return toDomainExpense(row), nil
}

func (r *ExpenseRepository) ListExpenses(ctx context.Context, filter interfaces.ExpenseListFilter) ([]*domain.Expense, error) {
	params := postgresdb.ListExpensesParams{
		UserID: filter.UserID,
		PageLimit:  int32(filter.Limit),
		PageOffset: int32(filter.Offset),
	}

	if filter.Category != nil {
		params.Category = sql.NullString{String: string(*filter.Category), Valid: true}
	} else {
		params.Category = sql.NullString{}
	}

	if filter.FromDate != nil {
		params.FromDate = sql.NullTime{Time: *filter.FromDate, Valid: true}
	}

	if filter.ToDate != nil {
		params.ToDate = sql.NullTime{Time: *filter.ToDate, Valid: true}
	}

	rows, err := r.q.ListExpenses(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}

	return toDomainExpenses(rows), nil
}

func (r *ExpenseRepository) UpdateExpense(ctx context.Context, expense *domain.Expense) error {
	return r.q.UpdateExpense(ctx, postgresdb.UpdateExpenseParams{
		ID:          expense.ID,
		UserID:      expense.UserID,
		Amount:      expense.Amount,
		Currency:    expense.Currency,
		Category:    string(expense.Category),
		Description: expense.Description,
		Date:        expense.Date,
		UpdatedAt:   expense.UpdatedAt,
	})
}

func (r *ExpenseRepository) DeleteExpense(ctx context.Context, userID string, expenseID string) error {
	if err := r.q.DeleteExpense(ctx, postgresdb.DeleteExpenseParams{
		ID:     expenseID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	return nil
}
