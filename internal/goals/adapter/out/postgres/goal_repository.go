package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

type GoalRepository struct {
	q postgresdb.Querier
}

var _ interfaces.GoalRepository = (*GoalRepository)(nil)

func NewGoalRepository(q postgresdb.Querier) *GoalRepository {
	return &GoalRepository{q: q}
}

func (r *GoalRepository) CreateGoal(ctx context.Context, goal *domain.Goal) (*domain.Goal, error) {
	if goal == nil {
		return nil, errors.New("goal is required")
	}

	arg := postgresdb.CreateGoalParams{
		ID:           goal.ID,
		UserID:       goal.UserID,
		Name:         goal.Name,
		Currency:     goal.Currency,
		TargetAmount: goal.TargetAmount,
		TargetDate: sql.NullTime{
			Time:  derefTime(goal.TargetDate),
			Valid: goal.TargetDate != nil,
		},
		Status:      string(goal.Status),
		CreatedAt:   goal.CreatedAt,
		UpdatedAt:   goal.UpdatedAt,
		CompletedAt: toNullTime(goal.CompletedAt),
	}

	row, err := r.q.CreateGoal(ctx, arg)
	if err != nil {
		return nil, err
	}

	return mapSQLGoalToDomain(&row), nil
}

func (r *GoalRepository) GetGoalByID(ctx context.Context, userID, goalID string) (*domain.Goal, error) {
	row, err := r.q.GetGoalByID(ctx, postgresdb.GetGoalByIDParams{
		ID:     goalID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapSQLGoalToDomain(&row), nil
}

func (r *GoalRepository) ListGoals(ctx context.Context, filter interfaces.GoalListFilter) ([]*domain.Goal, error) {
	var status sql.NullString
	if filter.Status != nil {
		status = sql.NullString{
			String: string(*filter.Status),
			Valid:  true,
		}
	}

	limit := filter.PageSize
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	args := postgresdb.ListGoalsParams{
		UserID:     filter.UserID,
		Status:     status,
		PageLimit:  int32(limit),
		PageOffset: int32(offset),
	}

	rows, err := r.q.ListGoals(ctx, args)
	if err != nil {
		return nil, err
	}

	out := make([]*domain.Goal, 0, len(rows))
	for i := range rows {
		out = append(out, mapSQLGoalToDomain(&rows[i]))
	}
	return out, nil
}

func (r *GoalRepository) UpdateGoal(ctx context.Context, goal *domain.Goal) error {
	if goal == nil {
		return errors.New("goal is required")
	}

	arg := postgresdb.UpdateGoalParams{
		ID:           goal.ID,
		UserID:       goal.UserID,
		Name:         goal.Name,
		TargetAmount: goal.TargetAmount,
		TargetDate: sql.NullTime{
			Time:  derefTime(goal.TargetDate),
			Valid: goal.TargetDate != nil,
		},
		Status:      string(goal.Status),
		UpdatedAt:   goal.UpdatedAt,
		CompletedAt: toNullTime(goal.CompletedAt),
	}

	return r.q.UpdateGoal(ctx, arg)
}

func (r *GoalRepository) DeleteGoal(ctx context.Context, userID, goalID string) error {
	err := r.q.DeleteGoal(ctx, postgresdb.DeleteGoalParams{
		ID:     goalID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if err == sql.ErrNoRows {
		return usecase.ErrGoalNotFound
	}
	return nil
}

func mapSQLGoalToDomain(row *postgresdb.Goal) *domain.Goal {
	var targetDate *time.Time
	if row.TargetDate.Valid {
		t := row.TargetDate.Time
		targetDate = &t
	}

	var completedAt *time.Time
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		completedAt = &t
	}

	return &domain.Goal{
		ID:           row.ID,
		UserID:       row.UserID,
		Name:         row.Name,
		Currency:     row.Currency,
		TargetAmount: row.TargetAmount,
		TargetDate:   targetDate,
		Status:       domain.GoalStatus(row.Status),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		CompletedAt:  completedAt,
	}
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}
