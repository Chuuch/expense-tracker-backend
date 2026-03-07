package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	repo "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func testDBURL(t *testing.T) string {
	t.Helper()

	v := os.Getenv("MIGRATE_DB_URL")
	if v == "" {
		t.Fatal("MIGRATE_DB_URL is not set")
	}
	return v
}

func newTestRepo(t *testing.T) (*repo.UserRepository, *sql.DB) {
	t.Helper()

	db, err := sql.Open("pgx", testDBURL(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	r := repo.NewUserRepository(postgresdb.New(db))
	return r, db
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func newDomainUser(email string) *domain.User {
	return domain.NewUser(
		fmt.Sprintf("01TEST%020d", time.Now().UnixNano()%1_000_000_000_000_000_000), // ULID-like unique id for tests
		email,
		"hashed-password",
		"Test",
		"+15550001111",
		"123 Main",
		"Austin",
		"TX",
		"78701",
		"US",
	)
}

func TestUserRepository_Create_And_GetByEmail(t *testing.T) {
	r, db := newTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	email := uniqueEmail("create_get")
	u := newDomainUser(email)

	created, err := r.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created user with id")
	}

	got, err := r.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}
	if got == nil {
		t.Fatalf("expected user, got nil")
	}
	if got.Email != email {
		t.Fatalf("expected email %s, got %s", email, got.Email)
	}
}

func TestUserRepository_GetByID_NotFoundReturnsNil(t *testing.T) {
	r, db := newTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	got, err := r.GetByID(ctx, "01NONEXISTENTUSERIDXXXXXXXXX")
	if err != nil {
		t.Fatalf("GetByID returned unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil user for missing id, got %+v", got)
	}
}

func TestUserRepository_UpdateUser_PersistsChanges(t *testing.T) {
	r, db := newTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	email := uniqueEmail("update")
	u := newDomainUser(email)

	created, err := r.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	created.Profile.Username = "Updated"
	created.Profile.City = "Dallas"
	created.UpdatedAt = time.Now()

	if err := r.UpdateUser(ctx, created); err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	got, err := r.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got == nil {
		t.Fatalf("expected user after update")
	}
	if got.Profile.Username != "Updated" || got.Profile.City != "Dallas" {
		t.Fatalf("update not persisted: %+v", got.Profile)
	}
}

func TestUserRepository_DeleteUser_SoftDeleteBehavior(t *testing.T) {
	r, db := newTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	email := uniqueEmail("delete")
	u := newDomainUser(email)

	created, err := r.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := r.DeleteUser(ctx, created.ID); err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	byID, err := r.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID after delete failed: %v", err)
	}
	if byID != nil {
		t.Fatalf("expected nil after delete, got %+v", byID)
	}

	byEmail, err := r.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetByEmail after delete failed: %v", err)
	}
	if byEmail != nil {
		t.Fatalf("expected nil by email after delete, got %+v", byEmail)
	}
}
