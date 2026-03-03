package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces"
)

type AnalyticsUsecase struct {
	repo interfaces.AnalyticsRepository
}

var _ interfaces.AnalyticsUsecase = (*AnalyticsUsecase)(nil)

func NewAnalyticsUsecase(repo interfaces.AnalyticsRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo}
}

func (u *AnalyticsUsecase) GetMonthlySpending(
	ctx context.Context,
	filter interfaces.MonthlySpendingFilter,
) ([]interfaces.MonthlySpendingPoint, error) {
	if filter.UserID == "" {
		return nil, fmt.Errorf("GetMonthlySpendin: user id is required")
	}

	if filter.From.IsZero() || filter.To.IsZero() {
		return nil, fmt.Errorf("GetMonthlySpending: from/to are required")
	}

	if filter.From.After(filter.To) {
		return nil, fmt.Errorf("GetMonthlySpending: from must be <= to")
	}

	filter.From = monthStart(filter.From)
	filter.To = monthStart(filter.To)

	points, err := u.repo.GetMonthlySpending(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("GetMonthlySpending: %w", err)
	}
	return points, nil
}

func (u *AnalyticsUsecase) GetSpendingByCategory(
	ctx context.Context,
	filter interfaces.CategorySpendingFilter,
) ([]interfaces.CategorySpendingPoint, error) {
	if filter.UserID == "" {
		return nil, fmt.Errorf("GetSpendingByCategory: user id is required")
	}
	if filter.FromDate.IsZero() || filter.ToDate.IsZero() {
		return nil, fmt.Errorf("GetSpendingByCategory: from/to are required")
	}
	if filter.FromDate.After(filter.ToDate) {
		return nil, fmt.Errorf("GetSpendingByCategory:from must be <= to")
	}

	points, err := u.repo.GetSpendingByCategory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("GetSpendingByCategory: %w", err)
	}
	return points, nil
}

func monthStart(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}
