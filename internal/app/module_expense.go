package app

import (
	expensehttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	expensepostgres "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/out/postgres"
	expenseusecase "github.com/chuuch/expense-tracker-backend/internal/expenses/usecase"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

func buildExpenseModule(q postgresdb.Querier) *expensehttp.ExpenseHandler {
	expenseRepo := expensepostgres.NewExpenseRepository(q)
	expenseUC := expenseusecase.NewExpenseUsecase(expenseRepo)
	return expensehttp.NewExpenseHandler(expenseUC)
}
