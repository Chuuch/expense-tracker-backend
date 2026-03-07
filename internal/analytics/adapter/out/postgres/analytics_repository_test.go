package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	analyticsrepo "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/out/postgres"
	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces"
	authrepo "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	authdomain "github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	expenserepo "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/out/postgres"
	expensedomain "github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func analyticsTestDBURL(t *testing.T) string {
	t.Helper()
	v := os.Getenv("MIGRATE_DB_URL")
	if v == "" {
		t.Fatal("MIGRATE_DB_URL is not set")
	}
	return v
}

func newAnalyticsTestRepos(t *testing.T) (*analyticsrepo.AnalyticsRepository, *authrepo.UserRepository, *expenserepo.ExpenseRepository, *sql.DB) {
	t.Helper()

	db, err := sql.Open("pgx", analyticsTestDBURL(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	q := postgresdb.New(db)
	return analyticsrepo.NewAnalyticsRepository(q), authrepo.NewUserRepository(q), expenserepo.NewExpenseRepository(q), db
}

func uniqueAnalyticsEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func newTestUserForAnalytics(t *testing.T, userRepo *authrepo.UserRepository, ctx context.Context) *authdomain.User {
	t.Helper()

	u := authdomain.NewUser(
		fmt.Sprintf("01ANAL%020d", time.Now().UnixNano()%1_000_000_000_000_000_000),
		uniqueAnalyticsEmail("analytics_repo"),
		"hashed-password",
		"Analytics",
		"Tester",
		"+15550004444",
		"123 Main",
		"Austin",
		"TX",
		"78701",
	)

	created, err := userRepo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("create user for analytics tests: %v", err)
	}
	return created
}

func monthStart(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

func TestAnalyticsRepository_GetMonthlySpending_AggregatesByMonthAndCurrency(t *testing.T) {
	analyticsRepo, userRepo, expenseRepo, db := newAnalyticsTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForAnalytics(t, userRepo, ctx)

	jan1 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	jan2 := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)

	exp1 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_anal_1_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      1000,
		Currency:    "USD",
		Category:    expensedomain.CategoryFood,
		Description: "Jan food",
		Date:        jan1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	exp2 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_anal_2_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      500,
		Currency:    "USD",
		Category:    expensedomain.CategoryTransport,
		Description: "Jan transport",
		Date:        jan2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	exp3 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_anal_3_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      2000,
		Currency:    "USD",
		Category:    expensedomain.CategoryHousing,
		Description: "Feb rent",
		Date:        feb1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, e := range []*expensedomain.Expense{exp1, exp2, exp3} {
		if _, err := expenseRepo.CreateExpense(ctx, e); err != nil {
			t.Fatalf("CreateExpense failed: %v", err)
		}
	}

	from := monthStart(jan1)
	to := monthStart(feb1)

	points, err := analyticsRepo.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: user.ID,
		From:   from,
		To:     to,
	})
	if err != nil {
		t.Fatalf("GetMonthlySpending failed: %v", err)
	}

	if len(points) != 2 {
		t.Fatalf("expected 2 monthly points (Jan + Feb), got %d: %+v", len(points), points)
	}
	if points[0].YearMonth != "2025-01" || points[0].Amount != 1500 || points[0].Currency != "USD" {
		t.Fatalf("expected first point 2025-01 1500 USD, got %+v", points[0])
	}
	if points[1].YearMonth != "2025-02" || points[1].Amount != 2000 || points[1].Currency != "USD" {
		t.Fatalf("expected second point 2025-02 2000 USD, got %+v", points[1])
	}
}

func TestAnalyticsRepository_GetMonthlySpending_EmptyRange_ReturnsEmpty(t *testing.T) {
	analyticsRepo, userRepo, _, db := newAnalyticsTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForAnalytics(t, userRepo, ctx)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	points, err := analyticsRepo.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: user.ID,
		From:   from,
		To:     to,
	})
	if err != nil {
		t.Fatalf("GetMonthlySpending failed: %v", err)
	}
	if len(points) != 0 {
		t.Fatalf("expected no points for range with no expenses, got %d", len(points))
	}
}

func TestAnalyticsRepository_GetSpendingByCategory_AggregatesAndOrdersByAmountDesc(t *testing.T) {
	analyticsRepo, userRepo, expenseRepo, db := newAnalyticsTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForAnalytics(t, userRepo, ctx)

	base := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)

	exp1 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_cat_1_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      3000,
		Currency:    "USD",
		Category:    expensedomain.CategoryHousing,
		Description: "Rent",
		Date:        base,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	exp2 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_cat_2_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      500,
		Currency:    "USD",
		Category:    expensedomain.CategoryFood,
		Description: "Groceries",
		Date:        base.Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	exp3 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_cat_3_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      700,
		Currency:    "USD",
		Category:    expensedomain.CategoryFood,
		Description: "Restaurant",
		Date:        base.Add(48 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	exp4 := &expensedomain.Expense{
		ID:          fmt.Sprintf("exp_cat_4_%d", time.Now().UnixNano()),
		UserID:      user.ID,
		Amount:      200,
		Currency:    "USD",
		Category:    expensedomain.CategoryTransport,
		Description: "Bus",
		Date:        base.Add(72 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, e := range []*expensedomain.Expense{exp1, exp2, exp3, exp4} {
		if _, err := expenseRepo.CreateExpense(ctx, e); err != nil {
			t.Fatalf("CreateExpense failed: %v", err)
		}
	}

	from := base.Add(-24 * time.Hour)
	to := base.Add(96 * time.Hour)

	points, err := analyticsRepo.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   user.ID,
		FromDate: from,
		ToDate:   to,
	})
	if err != nil {
		t.Fatalf("GetSpendingByCategory failed: %v", err)
	}

	if len(points) != 3 {
		t.Fatalf("expected 3 category points, got %d: %+v", len(points), points)
	}
	if points[0].Category != "housing" || points[0].Amount != 3000 {
		t.Fatalf("expected first category housing 3000, got %+v", points[0])
	}
	if points[1].Category != "food" || points[1].Amount != 1200 {
		t.Fatalf("expected second category food 1200, got %+v", points[1])
	}
	if points[2].Category != "transport" || points[2].Amount != 200 {
		t.Fatalf("expected third category transport 200, got %+v", points[2])
	}
	for _, p := range points {
		if p.Currency != "USD" {
			t.Fatalf("expected currency USD, got %s", p.Currency)
		}
	}
}

func TestAnalyticsRepository_GetSpendingByCategory_EmptyRange_ReturnsEmpty(t *testing.T) {
	analyticsRepo, userRepo, _, db := newAnalyticsTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForAnalytics(t, userRepo, ctx)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

	points, err := analyticsRepo.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   user.ID,
		FromDate: from,
		ToDate:   to,
	})
	if err != nil {
		t.Fatalf("GetSpendingByCategory failed: %v", err)
	}
	if len(points) != 0 {
		t.Fatalf("expected no category points for range with no expenses, got %d", len(points))
	}
}
