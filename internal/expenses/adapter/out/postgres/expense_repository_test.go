package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	authrepo "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	authdomain "github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	expenserepo "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/out/postgres"
	expensedomain "github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var expenseIDSeq uint64

func expenseTestDBURL(t *testing.T) string {
	t.Helper()
	v := os.Getenv("MIGRATE_DB_URL")
	if v == "" {
		t.Fatal("MIGRATE_DB_URL is not set")
	}
	return v
}

func newExpenseTestRepos(t *testing.T) (*expenserepo.ExpenseRepository, *authrepo.UserRepository, *sql.DB) {
	t.Helper()

	db, err := sql.Open("pgx", expenseTestDBURL(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	q := postgresdb.New(db)
	return expenserepo.NewExpenseRepository(q), authrepo.NewUserRepository(q), db
}

func uniqueExpenseID() string {
	n := atomic.AddUint64(&expenseIDSeq, 1)
	return fmt.Sprintf("exp_%d_%d", time.Now().UnixNano(), n)
}

func uniqueExpenseEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func newTestUserForExpenses(t *testing.T, userRepo *authrepo.UserRepository, ctx context.Context) *authdomain.User {
	t.Helper()

	u := authdomain.NewUser(
		fmt.Sprintf("01TEST%020d", time.Now().UnixNano()%1_000_000_000_000_000_000),
		uniqueExpenseEmail("expense_repo"),
		"hashed-password",
		"Expense",
	)

	created, err := userRepo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("create user for expense tests: %v", err)
	}
	return created
}

func newDomainExpense(userID string, amount int64, category expensedomain.ExpenseCategory, desc string, date time.Time) *expensedomain.Expense {
	return &expensedomain.Expense{
		ID:          uniqueExpenseID(),
		UserID:      userID,
		Amount:      amount,
		Currency:    "USD",
		Category:    category,
		Description: desc,
		Date:        date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestExpenseRepository_CreateExpense_And_GetByID(t *testing.T) {
	r, userRepo, db := newExpenseTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForExpenses(t, userRepo, ctx)

	exp := newDomainExpense(user.ID, 12500, expensedomain.CategoryHousing, "Monthly rent", time.Now().Add(-2*time.Hour))

	created, err := r.CreateExpense(ctx, exp)
	if err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}
	if created == nil {
		t.Fatal("expected created expense, got nil")
	}
	if created.ID != exp.ID || created.UserID != user.ID {
		t.Fatalf("unexpected created expense: %+v", created)
	}

	got, err := r.GetExpenseByID(ctx, user.ID, exp.ID)
	if err != nil {
		t.Fatalf("GetExpenseByID failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected expense, got nil")
	}
	if got.Amount != 12500 || got.Category != expensedomain.CategoryHousing || got.Description != "Monthly rent" {
		t.Fatalf("unexpected expense values: %+v", got)
	}
}

func TestExpenseRepository_ListExpenses_ByUserAndCategory(t *testing.T) {
	r, userRepo, db := newExpenseTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForExpenses(t, userRepo, ctx)

	// Same user, different categories
	exp1 := newDomainExpense(user.ID, 5000, expensedomain.CategoryFood, "Groceries", time.Now().Add(-3*time.Hour))
	exp2 := newDomainExpense(user.ID, 7000, expensedomain.CategoryHousing, "Utility bill", time.Now().Add(-2*time.Hour))
	exp3 := newDomainExpense(user.ID, 3000, expensedomain.CategoryFood, "Lunch", time.Now().Add(-1*time.Hour))

	for _, e := range []*expensedomain.Expense{exp1, exp2, exp3} {
		if _, err := r.CreateExpense(ctx, e); err != nil {
			t.Fatalf("CreateExpense failed: %v", err)
		}
	}

	food := expensedomain.CategoryFood
	items, err := r.ListExpenses(ctx, interfaces.ExpenseListFilter{
		UserID:   user.ID,
		Category: &food,
		Limit:    20,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListExpenses failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 food expenses, got %d", len(items))
	}
	for _, it := range items {
		if it.Category != expensedomain.CategoryFood {
			t.Fatalf("expected only food category, got %s", it.Category)
		}
	}
}

func TestExpenseRepository_DeleteExpense_RemovesRow(t *testing.T) {
	r, userRepo, db := newExpenseTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForExpenses(t, userRepo, ctx)

	exp := newDomainExpense(user.ID, 4200, expensedomain.CategoryTransport, "Taxi", time.Now())
	if _, err := r.CreateExpense(ctx, exp); err != nil {
		t.Fatalf("CreateExpense failed: %v", err)
	}

	if err := r.DeleteExpense(ctx, user.ID, exp.ID); err != nil {
		t.Fatalf("DeleteExpense failed: %v", err)
	}

	got, err := r.GetExpenseByID(ctx, user.ID, exp.ID)
	if err != nil {
		t.Fatalf("GetExpenseByID after delete failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}
