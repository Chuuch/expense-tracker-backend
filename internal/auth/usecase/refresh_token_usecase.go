package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/chuuch/expense-tracker-backend/utils"
)

var (
	ErrRefreshTokenInvalid = domain.ErrRefreshTokenInvalid
	ErrRefreshTokenExpired = domain.ErrRefreshTokenExpired
	ErrRefreshTokenRevoked = domain.ErrRefreshTokenRevoked
)

type RefreshTokenUsecase struct {
	refreshTokenRepo interfaces.RefreshTokenRepository
	userRepo         interfaces.UserRepository
	tokenUsecase     interfaces.TokenUsecase
	cfg              *config.Config
}

func NewRefreshTokenUsecase(
	refreshTokenRepo interfaces.RefreshTokenRepository,
	userRepo interfaces.UserRepository,
	tokenUsecase interfaces.TokenUsecase,
	cfg *config.Config,
) *RefreshTokenUsecase {
	return &RefreshTokenUsecase{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		tokenUsecase:     tokenUsecase,
		cfg:              cfg,
	}
}

func (u *RefreshTokenUsecase) IssueTokenPair(ctx context.Context, user *domain.User) (*domain.TokenPair, error) {
	if user == nil {
		return nil, fmt.Errorf("isssue token pair: user is required")
	}

	accessToken, err := u.tokenUsecase.GenerateToken(user, u.cfg.Auth.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("issue token pair: %w", err)
	}

	refreshToken, _, err := u.createAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *RefreshTokenUsecase) Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return nil, ErrRefreshTokenInvalid
	}

	hashed := hashToken(rawRefreshToken)
	stored, err := u.refreshTokenRepo.GetByHash(ctx, hashed)
	if err != nil {
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}

	if stored == nil {
		return nil, ErrRefreshTokenInvalid
	}

	if stored.IsRevoked() {
		return nil, ErrRefreshTokenRevoked
	}

	if stored.IsExpired(time.Now()) {
		return nil, ErrRefreshTokenExpired
	}

	user, err := u.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user by id for refresh: %w", err)
	}

	if user == nil {
		return nil, ErrRefreshTokenInvalid
	}

	accessToken, err := u.tokenUsecase.GenerateToken(user, u.cfg.Auth.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate access token from refresh: %w", err)
	}

	newRefreshToken, newRefreshID, err := u.createAndStoreRefreshToken(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.refreshTokenRepo.Revoke(ctx, stored.ID, &newRefreshID); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *RefreshTokenUsecase) Revoke(ctx context.Context, rawRefreshToken string) error {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return ErrRefreshTokenInvalid
	}

	hashed := hashToken(rawRefreshToken)
	stored, err := u.refreshTokenRepo.GetByHash(ctx, hashed)
	if err != nil {
		return fmt.Errorf("get refresh token for revoke: %w", err)
	}

	if stored == nil || stored.IsRevoked() {
		return nil
	}

	if err := u.refreshTokenRepo.Revoke(ctx, stored.ID, nil); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

func (u *RefreshTokenUsecase) DeleteExpired(ctx context.Context) error {
	if err := u.refreshTokenRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("delete expired refresh token: %w", err)
	}
	return nil
}

func (u *RefreshTokenUsecase) createAndStoreRefreshToken(ctx context.Context, userID string) (rawToken string, refreshID string, err error) {
	if strings.TrimSpace(userID) == "" {
		return "", "", errors.New("user id is required")
	}

	if u.cfg.Auth.RefreshTokenTTL <= 0 {
		return "", "", errors.New("refresh token ttl must be > 0")
	}

	raw, err := generateOpaqueToken(32)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	id := utils.GenerateULID()
	token := &domain.RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: hashToken(raw),
		ExpiresAt: now.Add(u.cfg.Auth.RefreshTokenTTL),
		CreatedAt: now,
	}

	if err := u.refreshTokenRepo.Create(ctx, token); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}

	return raw, id, nil
}

func generateOpaqueToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
