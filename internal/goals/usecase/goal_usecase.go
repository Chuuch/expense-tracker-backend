package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/utils"
)

type GoalUsecase struct {
	goalRepo         interfaces.GoalRepository
	contributionRepo interfaces.GoalContributionRepository
}

func NewGoalUsecase(
	goalRepo interfaces.GoalRepository,
	contributionRepo interfaces.GoalContributionRepository,
) *GoalUsecase {
	return &GoalUsecase{
		goalRepo:         goalRepo,
		contributionRepo: contributionRepo,
	}
}

var _ interfaces.GoalUsecase = (*GoalUsecase)(nil)

func (u *GoalUsecase) CreateGoal(
	ctx context.Context,
	userID string,
	name string,
	currency string,
	targetAmount int64,
	targetDate *time.Time,
) (*domain.Goal, error) {
	goal, err := domain.NewGoal(
		utils.GenerateULID(),
		userID,
		name,
		currency,
		targetAmount,
		targetDate,
	)
	if err != nil {
		return nil, fmt.Errorf("usecase.NewGoal: %w", err)
	}

	created, err := u.goalRepo.CreateGoal(ctx, goal)
	if err != nil {
		return nil, fmt.Errorf("usecase.CreateGoal: %w", err)
	}

	return created, nil
}

func (u *GoalUsecase) GetGoalByID(
	ctx context.Context,
	userID string,
	goalID string,
) (*domain.Goal, error) {
	goal, err := u.goalRepo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetGoalByID: %w", err)
	}
	if goal == nil {
		return nil, ErrGoalNotFound
	}
	return goal, nil
}

func (u *GoalUsecase) ListGoals(
	ctx context.Context,
	filter interfaces.GoalListFilter,
) ([]*domain.Goal, error) {
	if strings.TrimSpace(filter.UserID) == "" {
		return nil, ErrGoalInvalidFilter
	}

	if filter.Status != nil && !domain.IsValidGoalStatus(*filter.Status) {
		return nil, ErrGoalInvalidStatus
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	goals, err := u.goalRepo.ListGoals(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("usecase.ListGoals: %w", err)
	}
	return goals, nil
}

func (u *GoalUsecase) UpdateGoal(
	ctx context.Context,
	userID string,
	goalID string,
	name string,
	targetAmount int64,
	targetDate *time.Time,
	status domain.GoalStatus,
) (*domain.Goal, error) {
	existing, err := u.goalRepo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetGoalByID: %w", err)
	}
	if existing == nil {
		return nil, ErrGoalNotFound
	}

	if strings.TrimSpace(name) != "" {
		return nil, domain.ErrGoalNameRequired
	}

	if targetAmount <= 0 {
		return nil, domain.ErrGoalTargetAmountInvalid
	}

	if !domain.IsValidGoalStatus(status) {
		return nil, ErrGoalInvalidStatus
	}

	existing.Name = strings.TrimSpace(name)
	existing.TargetAmount = targetAmount
	existing.TargetDate = targetDate
	existing.UpdatedAt = time.Now()

	if existing.Status != status {
		if err := existing.TransitionTo(status, time.Now()); err != nil {
			return nil, fmt.Errorf("usecase.TransitionGoalStatus: %w", err)
		}
	}

	if err := u.goalRepo.UpdateGoal(ctx, existing); err != nil {
		return nil, fmt.Errorf("usecase.UpdateGoal: %w", err)
	}

	return existing, nil
}

func (u *GoalUsecase) DeleteGoal(
	ctx context.Context,
	userID string,
	goalID string,
) error {
	existing, err := u.goalRepo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return fmt.Errorf("usecase.GetGoalByID: %w", err)
	}
	if existing == nil {
		return ErrGoalNotFound
	}

	if err := u.goalRepo.DeleteGoal(ctx, userID, goalID); err != nil {
		return fmt.Errorf("usecase.DeleteGoal: %w", err)
	}
	return nil
}

func (u *GoalUsecase) AddContribution(
	ctx context.Context,
	userID string,
	goalID string,
	amount int64,
	contributionDate time.Time,
	note string,
) (*domain.GoalContribution, error) {
	goal, err := u.goalRepo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetGoalByID: %w", err)
	}

	if goal == nil {
		return nil, ErrGoalNotFound
	}

	switch goal.Status {
	case domain.GoalStatusCancelled:
		return nil, ErrGoalForbidden
	case domain.GoalStatusCompleted:
		return nil, ErrGoalAlreadyCompleted
	case domain.GoalStatusInactive:
		return nil, ErrGoalForbidden
	}

	contribution, err := domain.NewGoalContribution(
		utils.GenerateULID(),
		goalID,
		userID,
		amount,
		contributionDate,
		note,
	)

	if err != nil {
		return nil, fmt.Errorf("usecase.NewGoalContribution: %w", err)
	}

	created, err := u.contributionRepo.CreateContribution(ctx, contribution)
	if err != nil {
		return nil, fmt.Errorf("usecase.CreateContribution: %w", err)
	}

	total, err := u.contributionRepo.GetTotalContributedAmount(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetTotalContributionAmount: %w", err)
	}
	if total >= goal.TargetAmount && goal.Status != domain.GoalStatusCompleted {
		if err := goal.TransitionTo(domain.GoalStatusCompleted, time.Now()); err != nil {
			return nil, fmt.Errorf("usecase.CompleteGoalTransition: %w", err)
		}
		if err := u.goalRepo.UpdateGoal(ctx, goal); err != nil {
			return nil, fmt.Errorf("usecase.UpdateGoal: %w", err)
		}
	}
	return created, nil
}

func (u *GoalUsecase) ListContributions(
	ctx context.Context,
	userID string,
	goalID string,
	pageSize int,
	offset int,
) ([]*domain.GoalContribution, error) {
	goal, err := u.goalRepo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetGoalByID: %w", err)
	}
	if goal == nil {
		return nil, ErrGoalNotFound
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if offset < 0 {
		offset = 0
	}

	contributions, err := u.contributionRepo.ListContributions(
		ctx,
		userID,
		goalID,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("usecase.ListContributions: %w", err)
	}
	return contributions, nil
}

func (u *GoalUsecase) GetGoalProgress(
	ctx context.Context,
	userID string,
	goalID string,
) (*interfaces.GoalProgress, error) {
	goal, err := u.goalRepo.GetGoalByID(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetGoalByID: %w", err)
	}
	if goal == nil {
		return nil, ErrGoalNotFound
	}

	total, err := u.contributionRepo.GetTotalContributedAmount(ctx, userID, goalID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetTotalContributedAmount: %w", err)
	}

	remaining := goal.TargetAmount - total
	if remaining < 0 {
		remaining = 0
	}

	progress := 0.0
	if goal.TargetAmount > 0 {
		progress = (float64(total) / float64(goal.TargetAmount)) * 100
		if progress > 100 {
			progress = 100
		}
	}

	return &interfaces.GoalProgress{
		GoalID:             goal.ID,
		TargetAmount:       goal.TargetAmount,
		SavedAmount:        total,
		RemainingAmount:    remaining,
		ProgressPercentage: progress,
		IsCompleted:        total >= goal.TargetAmount || goal.Status == domain.GoalStatusCompleted,
		LastUpdatedAt:      goal.UpdatedAt,
	}, nil
}
