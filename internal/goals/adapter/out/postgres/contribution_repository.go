package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

type GoalContributionRepository struct {
	q postgresdb.Querier
}

var _ interfaces.GoalContributionRepository = (*GoalContributionRepository)(nil)

func NewGoalContributionRepository(q postgresdb.Querier) *GoalContributionRepository {
	return &GoalContributionRepository{q: q}
}

func (r *GoalContributionRepository) CreateContribution(
	ctx context.Context,
	contribution *domain.GoalContribution,
) (*domain.GoalContribution, error) {
	if contribution == nil {
		return nil, errors.New("contribution is required")
	}

	arg := postgresdb.CreateGoalContributionParams{
		ID:               contribution.ID,
		GoalID:           contribution.GoalID,
		UserID:           contribution.UserID,
		Amount:           contribution.Amount,
		ContributionDate: contribution.ContributionDate,
		Note: sql.NullString{
			String: contribution.Note,
			Valid:  contribution.Note != "",
		},
		CreatedAt: contribution.CreatedAt,
		UpdatedAt: contribution.UpdatedAt,
	}

	row, err := r.q.CreateGoalContribution(ctx, arg)
	if err != nil {
		return nil, err
	}

	return mapSQLContributionToDomain(&row), nil
}

func (r *GoalContributionRepository) ListContributions(
	ctx context.Context,
	userID string,
	goalID string,
	pageSize int,
	offset int,
) ([]*domain.GoalContribution, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.q.ListGoalContributions(ctx, postgresdb.ListGoalContributionsParams{
		UserID: userID,
		GoalID: goalID,
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	out := make([]*domain.GoalContribution, 0, len(rows))
	for i := range rows {
		out = append(out, mapSQLContributionToDomain(&rows[i]))
	}
	return out, nil
}

func (r *GoalContributionRepository) GetTotalContributedAmount(
	ctx context.Context,
	userID string,
	goalID string,
) (int64, error) {
	total, err := r.q.GetTotalContributedAmount(ctx, postgresdb.GetTotalContributedAmountParams{
		UserID: userID,
		GoalID: goalID,
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}

func mapSQLContributionToDomain(row *postgresdb.GoalContribution) *domain.GoalContribution {
	note := ""
	if row.Note.Valid {
		note = row.Note.String
	}
	return &domain.GoalContribution{
		ID:               row.ID,
		GoalID:           row.GoalID,
		UserID:           row.UserID,
		Amount:           row.Amount,
		ContributionDate: row.ContributionDate,
		Note:             note,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
