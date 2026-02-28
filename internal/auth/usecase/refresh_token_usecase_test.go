package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
)

type fakeRefreshTokenRepo struct {
	tokensByHash map[string]*domain.RefreshToken
	created      []*domain.RefreshToken
	revokedID    string
	replacedBy   *string
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{
		tokensByHash: make(map[string]*domain.RefreshToken),
		created:      make([]*domain.RefreshToken, 0),
	}
}

func (f *fakeRefreshTokenRepo) Create(_ context.Context, token *domain.RefreshToken) error {
	f.created = append(f.created, token)
	f.tokensByHash[token.TokenHash] = token
	return nil
}

func (f *fakeRefreshTokenRepo) GetByHash(_ context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if t, ok := f.tokensByHash[tokenHash]; ok {
		return t, nil
	}
	return nil, nil
}

func (f *fakeRefreshTokenRepo) Revoke(_ context.Context, id string, replacedBy *string) error {
	f.revokedID = id
	f.replacedBy = replacedBy

	for _, t := range f.tokensByHash {
		if t.ID == id {
			now := time.Now()
			t.RevokedAt = &now
			t.ReplacedBy = replacedBy
			break
		}
	}
	return nil
}

func (f *fakeRefreshTokenRepo) DeleteExpired(_ context.Context) error {
	return nil
}

type fakeUserRepo struct {
	user *domain.User
}

func (f *fakeUserRepo) CreateUser(_ context.Context, user *domain.User) (*domain.User, error) {
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if f.user != nil && f.user.ID == id {
		return f.user, nil
	}
	return nil, nil
}

func (f *fakeUserRepo) UpdateUser(_ context.Context, _ *domain.User) error {
	return nil
}

func (f *fakeUserRepo) DeleteUser(_ context.Context, _ string) error {
	return nil
}

type fakeTokenUsecase struct {
	token string
	err   error
}

func (f *fakeTokenUsecase) GenerateToken(_ *domain.User, _ time.Duration) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.token, nil
}

func (f *fakeTokenUsecase) VerifyToken(_ string) (*domain.TokenClaims, error) {
	return nil, errors.New("not implemented in test")
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func testRefreshUC(repo *fakeRefreshTokenRepo, userRepo *fakeUserRepo, tokenUC *fakeTokenUsecase) *usecase.RefreshTokenUsecase {
	return usecase.NewRefreshTokenUsecase(
		repo,
		userRepo,
		tokenUC,
		&config.Config{
			Auth: config.AuthConfig{
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 24 * time.Hour,
			},
		},
	)
}

func TestRefreshTokenUsecase_IssueTokenPair_Success(t *testing.T) {
	repo := newFakeRefreshTokenRepo()
	userRepo := &fakeUserRepo{}
	tokenUC := &fakeTokenUsecase{token: "access-token"}
	uc := testRefreshUC(repo, userRepo, tokenUC)

	user := &domain.User{ID: "user-1", Role: domain.RoleUser}
	pair, err := uc.IssueTokenPair(context.Background(), user)
	if err != nil {
		t.Fatalf("IssueTokenPair returned error: %v", err)
	}
	if pair == nil || pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected both access and refresh token, got %+v", pair)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one refresh token to be created, got %d", len(repo.created))
	}
}

func TestRefreshTokenUsecase_Refresh_UnknownToken(t *testing.T) {
	repo := newFakeRefreshTokenRepo()
	userRepo := &fakeUserRepo{user: &domain.User{ID: "user-1"}}
	tokenUC := &fakeTokenUsecase{token: "access-token"}
	uc := testRefreshUC(repo, userRepo, tokenUC)

	_, err := uc.Refresh(context.Background(), "unknown")
	if !errors.Is(err, usecase.ErrRefreshTokenInvalid) {
		t.Fatalf("expected ErrRefreshTokenInvalid, got %v", err)
	}
}

func TestRefreshTokenUsecase_Refresh_RevokedToken(t *testing.T) {
	repo := newFakeRefreshTokenRepo()
	userRepo := &fakeUserRepo{user: &domain.User{ID: "user-1"}}
	tokenUC := &fakeTokenUsecase{token: "access-token"}
	uc := testRefreshUC(repo, userRepo, tokenUC)

	now := time.Now()
	repo.tokensByHash[hash("rt-old")] = &domain.RefreshToken{
		ID:        "rt-id-1",
		UserID:    "user-1",
		TokenHash: hash("rt-old"),
		ExpiresAt: now.Add(1 * time.Hour),
		RevokedAt: &now,
		CreatedAt: now,
	}

	_, err := uc.Refresh(context.Background(), "rt-old")
	if !errors.Is(err, usecase.ErrRefreshTokenRevoked) {
		t.Fatalf("expected ErrRefreshTokenRevoked, got %v", err)
	}
}

func TestRefreshTokenUsecase_Refresh_SuccessRotatesAndRevokesOld(t *testing.T) {
	repo := newFakeRefreshTokenRepo()
	userRepo := &fakeUserRepo{user: &domain.User{ID: "user-1", Role: domain.RoleUser}}
	tokenUC := &fakeTokenUsecase{token: "access-new"}
	uc := testRefreshUC(repo, userRepo, tokenUC)

	now := time.Now()
	repo.tokensByHash[hash("rt-old")] = &domain.RefreshToken{
		ID:        "rt-id-old",
		UserID:    "user-1",
		TokenHash: hash("rt-old"),
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}

	pair, err := uc.Refresh(context.Background(), "rt-old")
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if pair == nil || pair.AccessToken != "access-new" || pair.RefreshToken == "" {
		t.Fatalf("unexpected token pair: %+v", pair)
	}
	if repo.revokedID != "rt-id-old" {
		t.Fatalf("expected old refresh token to be revoked, got revokedID=%s", repo.revokedID)
	}
	if repo.replacedBy == nil || *repo.replacedBy == "" {
		t.Fatalf("expected replacedBy to be set on old token")
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one new refresh token, got %d", len(repo.created))
	}
}
