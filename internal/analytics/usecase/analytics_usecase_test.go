package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces/mocks"
	"go.uber.org/mock/gomock"
)

func TestAnalyticsUsecase_GetMonthlySpending_MissingUserID_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	from := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	to := time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().GetMonthlySpending(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: "",
		From:   from,
		To:     to,
	})
	if err == nil {
		t.Fatal("expected error for missing user id")
	}
	if !contains(err.Error(), "user id") {
		t.Fatalf("expected error to mention user id, got %q", err.Error())
	}
}

func TestAnalyticsUsecase_GetMonthlySpending_ZeroFromOrTo_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().GetMonthlySpending(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: "u1",
		From:   time.Time{},
		To:     time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error for zero From")
	}
	if !contains(err.Error(), "from/to") {
		t.Fatalf("expected error to mention from/to, got %q", err.Error())
	}
}

func TestAnalyticsUsecase_GetMonthlySpending_FromAfterTo_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().GetMonthlySpending(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: "u1",
		From:   from,
		To:     to,
	})
	if err == nil {
		t.Fatal("expected error when from > to")
	}
	if !contains(err.Error(), "from must be <= to") {
		t.Fatalf("expected error to mention from <= to, got %q", err.Error())
	}
}

func TestAnalyticsUsecase_GetMonthlySpending_Success_NormalizesToMonthStartAndReturnsRepoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	from := time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC)
	to := time.Date(2025, 6, 20, 8, 0, 0, 0, time.UTC)

	repo.EXPECT().
		GetMonthlySpending(ctx, gomock.AssignableToTypeOf(interfaces.MonthlySpendingFilter{})).
		DoAndReturn(func(_ context.Context, f interfaces.MonthlySpendingFilter) ([]interfaces.MonthlySpendingPoint, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			// Usecase must normalize to start of month
			expectFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
			expectTo := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
			if !f.From.Equal(expectFrom) {
				t.Fatalf("expected From normalized to 2025-01-01, got %v", f.From)
			}
			if !f.To.Equal(expectTo) {
				t.Fatalf("expected To normalized to 2025-06-01, got %v", f.To)
			}
			return []interfaces.MonthlySpendingPoint{
				{YearMonth: "2025-01", Amount: 1500, Currency: "USD"},
				{YearMonth: "2025-02", Amount: 2000, Currency: "USD"},
			}, nil
		}).
		Times(1)

	points, err := uc.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: "u1",
		From:   from,
		To:     to,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if points[0].YearMonth != "2025-01" || points[0].Amount != 1500 {
		t.Fatalf("unexpected first point: %+v", points[0])
	}
	if points[1].YearMonth != "2025-02" || points[1].Amount != 2000 {
		t.Fatalf("unexpected second point: %+v", points[1])
	}
}

func TestAnalyticsUsecase_GetMonthlySpending_RepoError_WrapsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	repoErr := errors.New("db connection failed")
	repo.EXPECT().
		GetMonthlySpending(gomock.Any(), gomock.Any()).
		Return(nil, repoErr).
		Times(1)

	_, err := uc.GetMonthlySpending(ctx, interfaces.MonthlySpendingFilter{
		UserID: "u1",
		From:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func TestAnalyticsUsecase_GetSpendingByCategory_MissingUserID_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().GetSpendingByCategory(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   "",
		FromDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error for missing user id")
	}
	if !contains(err.Error(), "user id") {
		t.Fatalf("expected error to mention user id, got %q", err.Error())
	}
}

func TestAnalyticsUsecase_GetSpendingByCategory_ZeroFromOrTo_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().GetSpendingByCategory(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   "u1",
		FromDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Time{},
	})
	if err == nil {
		t.Fatal("expected error for zero ToDate")
	}
	if !contains(err.Error(), "from/to") {
		t.Fatalf("expected error to mention from/to, got %q", err.Error())
	}
}

func TestAnalyticsUsecase_GetSpendingByCategory_FromAfterTo_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	repo.EXPECT().GetSpendingByCategory(gomock.Any(), gomock.Any()).Times(0)

	_, err := uc.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   "u1",
		FromDate: from,
		ToDate:   to,
	})
	if err == nil {
		t.Fatal("expected error when from > to")
	}
	if !contains(err.Error(), "from must be <= to") {
		t.Fatalf("expected error to mention from <= to, got %q", err.Error())
	}
}

func TestAnalyticsUsecase_GetSpendingByCategory_Success_ReturnsRepoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	from := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)

	repo.EXPECT().
		GetSpendingByCategory(ctx, gomock.AssignableToTypeOf(interfaces.CategorySpendingFilter{})).
		DoAndReturn(func(_ context.Context, f interfaces.CategorySpendingFilter) ([]interfaces.CategorySpendingPoint, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if !f.FromDate.Equal(from) || !f.ToDate.Equal(to) {
				t.Fatalf("expected from/to unchanged, got %v / %v", f.FromDate, f.ToDate)
			}
			return []interfaces.CategorySpendingPoint{
				{Category: "housing", Amount: 3000, Currency: "USD"},
				{Category: "food", Amount: 1200, Currency: "USD"},
			}, nil
		}).
		Times(1)

	points, err := uc.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   "u1",
		FromDate: from,
		ToDate:   to,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if points[0].Category != "housing" || points[0].Amount != 3000 {
		t.Fatalf("unexpected first point: %+v", points[0])
	}
	if points[1].Category != "food" || points[1].Amount != 1200 {
		t.Fatalf("unexpected second point: %+v", points[1])
	}
}

func TestAnalyticsUsecase_GetSpendingByCategory_RepoError_WrapsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAnalyticsRepository(ctrl)
	uc := usecase.NewAnalyticsUsecase(repo)
	ctx := context.Background()

	repoErr := errors.New("db error")
	repo.EXPECT().
		GetSpendingByCategory(gomock.Any(), gomock.Any()).
		Return(nil, repoErr).
		Times(1)

	_, err := uc.GetSpendingByCategory(ctx, interfaces.CategorySpendingFilter{
		UserID:   "u1",
		FromDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ToDate:   time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
