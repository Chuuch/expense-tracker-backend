package interfaces

import (
	"context"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
)

type GoalListFilter struct {
	UserID   string
	Status   *domain.GoalStatus
	PageSize int
	Offset   int
}

type GoalProgress struct {
	GoalID             string
	TargetAmount       int64
	SavedAmount        int64
	RemainingAmount    int64
	ProgressPercentage float64
	IsCompleted        bool
	LastUpdatedAt      time.Time
}

type GoalUsecase interface {
	CreateGoal(
		ctx context.Context,
		userID string,
		name string,
		currency string,
		targetAmount int64,
		targetDate *time.Time,
	) (*domain.Goal, error)

	GetGoalByID(ctx context.Context, userID, goalID string) (*domain.Goal, error)
	ListGoals(ctx context.Context, filter GoalListFilter) ([]*domain.Goal, error)

	UpdateGoal(
		ctx context.Context,
		userID string,
		goalID string,
		name string,
		targetAmount int64,
		targetDate *time.Time,
		status domain.GoalStatus,
	) (*domain.Goal, error)

	DeleteGoal(ctx context.Context, userID, goalID string) error

	AddContribution(
		ctx context.Context,
		userID string,
		goalID string,
		amount int64,
		contributionDate time.Time,
		note string,
	) (*domain.GoalContribution, error)

	ListContributions(
		ctx context.Context,
		userID string,
		goalID string,
		pageSize int,
		offset int,
	) ([]*domain.GoalContribution, error)

	GetGoalProgress(ctx context.Context, userID, goalID string) (*GoalProgress, error)
}
