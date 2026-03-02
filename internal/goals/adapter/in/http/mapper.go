package http

import (
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
)

func mapGoalToResponse(goal *domain.Goal) *GoalResponse {
	var targetDate *string
	if goal.TargetDate != nil {
		td := goal.TargetDate.Format("2006-01-02")
		targetDate = &td
	}

	var completedAt *string
	if goal.CompletedAt != nil {
		ca := goal.CompletedAt.Format(time.RFC3339)
		completedAt = &ca
	}

	return &GoalResponse{
		ID:           goal.ID,
		UserID:       goal.UserID,
		Name:         goal.Name,
		TargetAmount: goal.TargetAmount,
		TargetDate:   targetDate,
		Status:       string(goal.Status),
		CreatedAt:    goal.CreatedAt,
		UpdatedAt:    goal.UpdatedAt,
		CompletedAt:  completedAt,
	}
}

func mapGoalsToResponse(goals []*domain.Goal, pageSize, offset int) ListGoalsResponse {
	responses := make([]GoalResponse, 0, len(goals))
	for _, goal := range goals {
		responses = append(responses, *mapGoalToResponse(goal))
	}
	return ListGoalsResponse{
		Goals:    responses,
		PageSize: pageSize,
		Offset:   offset,
	}
}

func mapContributionToResponse(c *domain.GoalContribution) GoalContributionResponse {
	return GoalContributionResponse{
		ID:               c.ID,
		GoalID:           c.GoalID,
		UserID:           c.UserID,
		Amount:           c.Amount,
		ContributionDate: c.ContributionDate.Format("2006-01-02"),
		Note:             c.Note,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

func mapContributionsToResponse(
	contributions []*domain.GoalContribution,
	pageSize int,
	offset int,
) ListContributionsResponse {
	items := make([]GoalContributionResponse, 0, len(contributions))
	for _, c := range contributions {
		items = append(items, mapContributionToResponse(c))
	}
	return ListContributionsResponse{
		Contributions: items,
		PageSize:      pageSize,
		Offset:        offset,
	}
}

func mapProgressToResponse(progress *interfaces.GoalProgress) GoalProgressResponse {
	return GoalProgressResponse{
		GoalID:             progress.GoalID,
		TargetAmount:       progress.TargetAmount,
		SavedAmount:        progress.SavedAmount,
		RemainingAmount:    progress.RemainingAmount,
		ProgressPercentage: progress.ProgressPercentage,
		IsCompleted:        progress.IsCompleted,
		LastUpdatedAt:      progress.LastUpdatedAt.Format(time.RFC3339),
	}
}
