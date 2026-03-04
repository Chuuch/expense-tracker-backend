package interfaces

import "context"

type AnalyticsRepository interface {
	GetMonthlySpending(ctx context.Context, filter MonthlySpendingFilter) ([]MonthlySpendingPoint, error)
	GetSpendingByCategory(ctx context.Context, filter CategorySpendingFilter) ([]CategorySpendingPoint, error)
}
