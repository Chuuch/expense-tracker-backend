package postgres_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	repo "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRefreshTestRepo(t *testing.T) (*repo.RefreshTokenRepository, *repo.UserRepository, *sql.DB) {
	t.Helper()

	db, err := sql.Open("pgx", testDBURL(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	q := postgresdb.New(db)
	return repo.NewRefreshTokenRepository(q), repo.NewUserRepository(q), db
}

func refreshTokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func refreshTokenID() string {
	return fmt.Sprintf("rt_%d", time.Now().UnixNano())
}

func createUserForRefreshTests(t *testing.T, userRepo *repo.UserRepository, ctx context.Context) *domain.User {
	t.Helper()

	user, err := userRepo.CreateUser(ctx, newDomainUser(uniqueEmail("refresh_repo")))
	if err != nil {
		t.Fatalf("create user for refresh tests: %v", err)
	}
	return user
}

func TestRefreshTokenRepository_Create_And_GetByHash(t *testing.T) {
	r, userRepo, db := newRefreshTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := createUserForRefreshTests(t, userRepo, ctx)

	raw := fmt.Sprintf("refresh-raw-2-%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(24 * time.Hour)
	token := &domain.RefreshToken{
		ID:        refreshTokenID(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash(raw),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := r.Create(ctx, token); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := r.GetByHash(ctx, refreshTokenHash(raw))
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	if got == nil {
		t.Fatalf("expected refresh token, got nil")
	}
	if got.ID != token.ID || got.UserID != user.ID || got.TokenHash != token.TokenHash {
		t.Fatalf("unexpected token: %+v", got)
	}
}

func TestRefreshTokenRepository_Revoke_SetsRevokedAtAndReplacedBy(t *testing.T) {
	r, userRepo, db := newRefreshTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := createUserForRefreshTests(t, userRepo, ctx)

	raw := fmt.Sprintf("refresh-raw-2-%d", time.Now().UnixNano())
	token := &domain.RefreshToken{
		ID:        refreshTokenID(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash(raw),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := r.Create(ctx, token); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	replacedBy := refreshTokenID()
	if err := r.Revoke(ctx, token.ID, &replacedBy); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	got, err := r.GetByHash(ctx, refreshTokenHash(raw))
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	if got == nil {
		t.Fatalf("expected refresh token, got nil")
	}
	if got.RevokedAt == nil {
		t.Fatalf("expected revoked_at to be set")
	}
	if got.ReplacedBy == nil || *got.ReplacedBy != replacedBy {
		t.Fatalf("expected replaced_by=%s, got %+v", replacedBy, got.ReplacedBy)
	}
}

func TestRefreshTokenRepository_DeleteExpired_RemovesOnlyExpired(t *testing.T) {
	r, userRepo, db := newRefreshTestRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := createUserForRefreshTests(t, userRepo, ctx)

	expiredRaw := fmt.Sprintf("refresh-expired-%d", time.Now().UnixNano())
	activeRaw := fmt.Sprintf("refresh-active-%d", time.Now().UnixNano())

	expired := &domain.RefreshToken{
		ID:        refreshTokenID(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash(expiredRaw),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	active := &domain.RefreshToken{
		ID:        refreshTokenID(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash(activeRaw),
		ExpiresAt: time.Now().Add(2 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := r.Create(ctx, expired); err != nil {
		t.Fatalf("Create expired failed: %v", err)
	}
	if err := r.Create(ctx, active); err != nil {
		t.Fatalf("Create active failed: %v", err)
	}

	if err := r.DeleteExpired(ctx); err != nil {
		t.Fatalf("DeleteExpired failed: %v", err)
	}

	gotExpired, err := r.GetByHash(ctx, refreshTokenHash(expiredRaw))
	if err != nil {
		t.Fatalf("GetByHash(expired) failed: %v", err)
	}
	if gotExpired != nil {
		t.Fatalf("expected expired token to be deleted, got %+v", gotExpired)
	}

	gotActive, err := r.GetByHash(ctx, refreshTokenHash(activeRaw))
	if err != nil {
		t.Fatalf("GetByHash(active) failed: %v", err)
	}
	if gotActive == nil {
		t.Fatalf("expected active token to remain")
	}
}
