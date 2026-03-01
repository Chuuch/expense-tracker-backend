package interfaces

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
)

type GoalRepository interface {
	CreateGoal(ctx context.Context, goal *domain.Goal) (*domain.Goal, error)
	GetGoalByID(ctx context.Context, userID, goalID string) (*domain.Goal, error)
	ListGoals(ctx context.Context, filter GoalListFilter) ([]*domain.Goal, error)
	UpdateGoal(ctx context.Context, goal *domain.Goal) error
	DeleteGoal(ctx context.Context, userID, goalID string) error
}

type GoalContributionRepository interface {
	CreateContribution(ctx context.Context, contribution *domain.GoalContribution) (*domain.GoalContribution, error)
	ListContributions(ctx context.Context, userID, goalID string, pageSize, offset int) ([]*domain.GoalContribution, error)
	GetTotalContributedAmount(ctx context.Context, userID, goalID string) (int64, error)
}
