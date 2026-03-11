package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

type RefreshTokenRepository struct {
	q postgresdb.Querier
}

func NewRefreshTokenRepository(q postgresdb.Querier) *RefreshTokenRepository {
	return &RefreshTokenRepository{q: q}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	if token == nil {
		return fmt.Errorf("create refreshtoken: token is nil")
	}

	return r.q.CreateRefreshToken(ctx, postgresdb.CreateRefreshTokenParams{
		ID:         token.ID,
		UserID:     token.UserID,
		TokenHash:  token.TokenHash,
		ExpiresAt:  token.ExpiresAt,
		RevokedAt:  toNullTime(token.RevokedAt),
		ReplacedBy: toNullString(token.ReplacedBy),
		CreatedAt:  token.CreatedAt,
	})
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}

	return toDomainRefreshToken(row), nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, replacedBy *string) error {
	return r.q.RevokeRefreshToken(ctx, postgresdb.RevokeRefreshTokenParams{
		ID:         id,
		ReplacedBy: toNullString(replacedBy),
	})
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.q.DeleteExpiredRefreshTokens(ctx)
}

func toDomainRefreshToken(t postgresdb.RefreshToken) *domain.RefreshToken {
	var revokedAt *time.Time
	if t.RevokedAt.Valid {
		v := t.RevokedAt.Time
		revokedAt = &v
	}

	var replacedBy *string
	if t.ReplacedBy.Valid {
		v := t.ReplacedBy.String
		replacedBy = &v
	}

	return &domain.RefreshToken{
		ID:         t.ID,
		UserID:     t.UserID,
		TokenHash:  t.TokenHash,
		ExpiresAt:  t.ExpiresAt,
		RevokedAt:  revokedAt,
		ReplacedBy: replacedBy,
		CreatedAt:  t.CreatedAt,
	}
}

func toNullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}
