package postgres

import (
	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

func toDomainExpense(e postgresdb.Expense) *domain.Expense {
	return &domain.Expense{
		ID: e.ID,
		UserID: e.UserID,
		Amount: e.Amount,
		Currency: e.Currency,
		Category: domain.ExpenseCategory(e.Category),
		Description: e.Description,
		Date: e.Date,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func toDomainExpenses(es []postgresdb.Expense) []*domain.Expense {
	out := make([]*domain.Expense, 0, len(es))
	for _, e := range es {
		out = append(out, toDomainExpense(e))
	}
	return out
}