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
	goaldomain "github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	goalrepo "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/out/postgres"
	interfaces "github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var goalIDSeq uint64

func goalTestDBURL(t *testing.T) string {
	t.Helper()
	v := os.Getenv("MIGRATE_DB_URL")
	if v == "" {
		t.Fatal("MIGRATE_DB_URL is not set")
	}
	return v
}

func newGoalTestRepos(t *testing.T) (*goalrepo.GoalRepository, *authrepo.UserRepository, *sql.DB) {
	t.Helper()

	db, err := sql.Open("pgx", goalTestDBURL(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	q := postgresdb.New(db)
	return goalrepo.NewGoalRepository(q), authrepo.NewUserRepository(q), db
}

func uniqueGoalID() string {
	n := atomic.AddUint64(&goalIDSeq, 1)
	return fmt.Sprintf("goal_%d_%d", time.Now().UnixNano(), n)
}

func uniqueGoalEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func newTestUserForGoals(t *testing.T, userRepo *authrepo.UserRepository, ctx context.Context) *authdomain.User {
	t.Helper()

	u := authdomain.NewUser(
		fmt.Sprintf("01GOAL%020d", time.Now().UnixNano()%1_000_000_000_000_000_000),
		uniqueGoalEmail("goals_repo"),
		"hashed-password",
		"Goal",
		"Tester",
		"+15550003333",
		"123 Main",
		"Austin",
		"TX",
		"78701",
	)

	created, err := userRepo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("create user for goals tests: %v", err)
	}
	return created
}

func newDomainGoal(userID string, targetAmount int64, name string, targetDate *time.Time) *goaldomain.Goal {
	return &goaldomain.Goal{
		ID:           uniqueGoalID(),
		UserID:       userID,
		Name:         name,
		Currency:     "USD",
		TargetAmount: targetAmount,
		TargetDate:   targetDate,
		Status:       goaldomain.GoalStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestGoalRepository_CreateGoal_And_GetByID(t *testing.T) {
	r, userRepo, db := newGoalTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForGoals(t, userRepo, ctx)

	td := time.Now().Add(30 * 24 * time.Hour).UTC()
	g := newDomainGoal(user.ID, 100000, "Emergency Fund", &td)

	created, err := r.CreateGoal(ctx, g)
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}
	if created == nil {
		t.Fatal("expected created goal, got nil")
	}
	if created.ID != g.ID || created.UserID != user.ID {
		t.Fatalf("unexpected created goal: %+v", created)
	}

	got, err := r.GetGoalByID(ctx, user.ID, g.ID)
	if err != nil {
		t.Fatalf("GetGoalByID failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected goal, got nil")
	}
	if got.Name != "Emergency Fund" || got.TargetAmount != 100000 {
		t.Fatalf("unexpected goal values: %+v", got)
	}
}

func TestGoalRepository_ListGoals_ByUserAndStatus(t *testing.T) {
	r, userRepo, db := newGoalTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForGoals(t, userRepo, ctx)

	now := time.Now().UTC()
	activeGoal := newDomainGoal(user.ID, 50000, "Active Goal", &now)
	completedGoal := newDomainGoal(user.ID, 30000, "Completed Goal", nil)
	completedGoal.Status = goaldomain.GoalStatusCompleted

	for _, g := range []*goaldomain.Goal{activeGoal, completedGoal} {
		if _, err := r.CreateGoal(ctx, g); err != nil {
			t.Fatalf("CreateGoal failed: %v", err)
		}
	}

	status := goaldomain.GoalStatusActive
	items, err := r.ListGoals(ctx, interfaces.GoalListFilter{
		UserID:   user.ID,
		Status:   &status,
		PageSize: 20,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListGoals failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 active goal, got %d", len(items))
	}
	if items[0].Status != goaldomain.GoalStatusActive {
		t.Fatalf("expected active status, got %s", items[0].Status)
	}
}

func TestGoalRepository_DeleteGoal_RemovesRow(t *testing.T) {
	r, userRepo, db := newGoalTestRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForGoals(t, userRepo, ctx)

	g := newDomainGoal(user.ID, 42000, "To delete", nil)
	if _, err := r.CreateGoal(ctx, g); err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}

	if err := r.DeleteGoal(ctx, user.ID, g.ID); err != nil {
		t.Fatalf("DeleteGoal failed: %v", err)
	}

	got, err := r.GetGoalByID(ctx, user.ID, g.ID)
	if err != nil {
		t.Fatalf("GetGoalByID after delete failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}