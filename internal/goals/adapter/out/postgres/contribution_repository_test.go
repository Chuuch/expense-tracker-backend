package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	authrepo "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	authdomain "github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	goalrepo "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/out/postgres"
	goaldomain "github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newGoalAndContributionRepos(t *testing.T) (*goalrepo.GoalRepository, *goalrepo.GoalContributionRepository, *authrepo.UserRepository, *sql.DB) {
	t.Helper()

	db, err := sql.Open("pgx", goalTestDBURL(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	q := postgresdb.New(db)
	return goalrepo.NewGoalRepository(q), goalrepo.NewGoalContributionRepository(q), authrepo.NewUserRepository(q), db
}

func newTestUserForContributions(t *testing.T, userRepo *authrepo.UserRepository, ctx context.Context) *authdomain.User {
	t.Helper()
	return newTestUserForGoals(t, userRepo, ctx)
}

func TestGoalContributionRepository_Create_List_Total(t *testing.T) {
	goalRepo, contribRepo, userRepo, db := newGoalAndContributionRepos(t)
	defer db.Close()

	ctx := context.Background()
	user := newTestUserForContributions(t, userRepo, ctx)

	goal := newDomainGoal(user.ID, 100000, "Savings", nil)
	createdGoal, err := goalRepo.CreateGoal(ctx, goal)
	if err != nil {
		t.Fatalf("CreateGoal failed: %v", err)
	}

	// contributions
	c1, err := goaldomain.NewGoalContribution(
		uniqueGoalID(),
		createdGoal.ID,
		user.ID,
		25000,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		"first deposit",
	)
	if err != nil {
		t.Fatalf("NewGoalContribution failed: %v", err)
	}
	c2, err := goaldomain.NewGoalContribution(
		uniqueGoalID(),
		createdGoal.ID,
		user.ID,
		15000,
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		"second deposit",
	)
	if err != nil {
		t.Fatalf("NewGoalContribution failed: %v", err)
	}

	for _, c := range []*goaldomain.GoalContribution{c1, c2} {
		if _, err := contribRepo.CreateContribution(ctx, c); err != nil {
			t.Fatalf("CreateContribution failed: %v", err)
		}
	}

	items, err := contribRepo.ListContributions(ctx, user.ID, createdGoal.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListContributions failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 contributions, got %d", len(items))
	}

	total, err := contribRepo.GetTotalContributedAmount(ctx, user.ID, createdGoal.ID)
	if err != nil {
		t.Fatalf("GetTotalContributedAmount failed: %v", err)
	}
	if total != 40000 {
		t.Fatalf("expected total 40000, got %d", total)
	}
}
