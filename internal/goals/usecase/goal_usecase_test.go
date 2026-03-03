package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces/mocks"
	"go.uber.org/mock/gomock"
)

func TestGoalUsecase_CreateGoal_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)

	ctx := context.Background()
	targetDate := time.Now().Add(30 * 24 * time.Hour)

	goalRepo.EXPECT().
		CreateGoal(ctx, gomock.AssignableToTypeOf(&domain.Goal{})).
		DoAndReturn(func(_ context.Context, g *domain.Goal) (*domain.Goal, error) {
			if g.ID == "" {
				t.Fatalf("expected generated goal ID")
			}
			if g.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", g.UserID)
			}
			if g.Name != "Emergency Fund" {
				t.Fatalf("expected name Emergency Fund, got %s", g.Name)
			}
			if g.Currency != "USD" {
				t.Fatalf("expected normalized currency USD, got %s", g.Currency)
			}
			if g.TargetAmount != 100000 {
				t.Fatalf("expected target 100000, got %d", g.TargetAmount)
			}
			if g.Status != domain.GoalStatusActive {
				t.Fatalf("expected status active, got %s", g.Status)
			}
			if g.TargetDate == nil {
				t.Fatalf("expected target date to be set")
			}
			return g, nil
		}).
		Times(1)

	goal, err := uc.CreateGoal(ctx, "u1", "Emergency Fund", "usd", 100000, &targetDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goal == nil {
		t.Fatal("expected goal, got nil")
	}
}

func TestGoalUsecase_GetGoalByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g_missing").
		Return(nil, nil).
		Times(1)

	_, err := uc.GetGoalByID(ctx, "u1", "g_missing")
	if !errors.Is(err, usecase.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}

func TestGoalUsecase_ListGoals_InvalidFilterMissingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().ListGoals(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.ListGoals(ctx, interfaces.GoalListFilter{
		UserID:   "",
		PageSize: 20,
		Offset:   0,
	})
	if !errors.Is(err, usecase.ErrGoalInvalidFilter) {
		t.Fatalf("expected ErrGoalInvalidFilter, got %v", err)
	}
}

func TestGoalUsecase_ListGoals_AppliesDefaultsAndCapsLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().
		ListGoals(ctx, gomock.AssignableToTypeOf(interfaces.GoalListFilter{})).
		DoAndReturn(func(_ context.Context, f interfaces.GoalListFilter) ([]*domain.Goal, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if f.PageSize != 100 {
				t.Fatalf("expected capped page size 100, got %d", f.PageSize)
			}
			if f.Offset != 0 {
				t.Fatalf("expected normalized offset 0, got %d", f.Offset)
			}
			return []*domain.Goal{}, nil
		}).
		Times(1)

	_, err := uc.ListGoals(ctx, interfaces.GoalListFilter{
		UserID:   "u1",
		PageSize: 999,
		Offset:   -10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGoalUsecase_DeleteGoal_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g_missing").
		Return(nil, nil).
		Times(1)

	goalRepo.EXPECT().
		DeleteGoal(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	err := uc.DeleteGoal(ctx, "u1", "g_missing")
	if !errors.Is(err, usecase.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}

func TestGoalUsecase_DeleteGoal_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g1").
		Return(&domain.Goal{ID: "g1", UserID: "u1"}, nil).
		Times(1)

	goalRepo.EXPECT().
		DeleteGoal(ctx, "u1", "g1").
		Return(nil).
		Times(1)

	if err := uc.DeleteGoal(ctx, "u1", "g1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGoalUsecase_AddContribution_GoalNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g_missing").
		Return(nil, nil).
		Times(1)

	contributionRepo.EXPECT().
		CreateContribution(gomock.Any(), gomock.Any()).
		Times(0)

	_, err := uc.AddContribution(ctx, "u1", "g_missing", 1000, time.Now(), "seed")
	if !errors.Is(err, usecase.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}

func TestGoalUsecase_AddContribution_InactiveGoal_ReturnsForbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g1").
		Return(&domain.Goal{
			ID:           "g1",
			UserID:       "u1",
			Status:       domain.GoalStatusInactive,
			TargetAmount: 10000,
		}, nil).
		Times(1)

	contributionRepo.EXPECT().
		CreateContribution(gomock.Any(), gomock.Any()).
		Times(0)

	_, err := uc.AddContribution(ctx, "u1", "g1", 500, time.Now(), "paused/inactive check")
	if !errors.Is(err, usecase.ErrGoalForbidden) {
		t.Fatalf("expected ErrGoalForbidden, got %v", err)
	}
}

func TestGoalUsecase_AddContribution_CompletesGoalWhenTargetReached(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	goal := &domain.Goal{
		ID:           "g1",
		UserID:       "u1",
		Name:         "Emergency Fund",
		Currency:     "USD",
		TargetAmount: 10000,
		Status:       domain.GoalStatusActive,
		CreatedAt:    time.Now().Add(-48 * time.Hour),
		UpdatedAt:    time.Now().Add(-24 * time.Hour),
	}

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g1").
		Return(goal, nil).
		Times(1)

	contributionRepo.EXPECT().
		CreateContribution(ctx, gomock.AssignableToTypeOf(&domain.GoalContribution{})).
		DoAndReturn(func(_ context.Context, c *domain.GoalContribution) (*domain.GoalContribution, error) {
			if c.ID == "" {
				t.Fatalf("expected generated contribution ID")
			}
			if c.Amount != 2500 {
				t.Fatalf("expected amount 2500, got %d", c.Amount)
			}
			return c, nil
		}).
		Times(1)

	contributionRepo.EXPECT().
		GetTotalContributedAmount(ctx, "u1", "g1").
		Return(int64(10000), nil).
		Times(1)

	goalRepo.EXPECT().
		UpdateGoal(ctx, gomock.AssignableToTypeOf(&domain.Goal{})).
		DoAndReturn(func(_ context.Context, g *domain.Goal) error {
			if g.Status != domain.GoalStatusCompleted {
				t.Fatalf("expected completed status after reaching target, got %s", g.Status)
			}
			if g.CompletedAt == nil {
				t.Fatalf("expected completed_at to be set")
			}
			return nil
		}).
		Times(1)

	_, err := uc.AddContribution(ctx, "u1", "g1", 2500, time.Now(), "final push")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGoalUsecase_GetGoalProgress_ComputesAmountsAndPercentage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalRepo := mocks.NewMockGoalRepository(ctrl)
	contributionRepo := mocks.NewMockGoalContributionRepository(ctrl)
	uc := usecase.NewGoalUsecase(goalRepo, contributionRepo)
	ctx := context.Background()

	updatedAt := time.Now().Add(-1 * time.Hour)

	goalRepo.EXPECT().
		GetGoalByID(ctx, "u1", "g1").
		Return(&domain.Goal{
			ID:           "g1",
			UserID:       "u1",
			TargetAmount: 20000,
			Status:       domain.GoalStatusActive,
			UpdatedAt:    updatedAt,
		}, nil).
		Times(1)

	contributionRepo.EXPECT().
		GetTotalContributedAmount(ctx, "u1", "g1").
		Return(int64(5000), nil).
		Times(1)

	progress, err := uc.GetGoalProgress(ctx, "u1", "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress == nil {
		t.Fatal("expected progress, got nil")
	}
	if progress.SavedAmount != 5000 {
		t.Fatalf("expected saved 5000, got %d", progress.SavedAmount)
	}
	if progress.RemainingAmount != 15000 {
		t.Fatalf("expected remaining 15000, got %d", progress.RemainingAmount)
	}
	if progress.ProgressPercentage != 25 {
		t.Fatalf("expected progress 25, got %v", progress.ProgressPercentage)
	}
	if progress.IsCompleted {
		t.Fatalf("expected not completed")
	}
	if !progress.LastUpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected LastUpdatedAt to match goal UpdatedAt")
	}
}
