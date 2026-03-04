package interfaces

import (
	"context"
	"time"
)

type MonthlySpendingPoint struct {
	YearMonth string `json:"year_month"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type CategorySpendingPoint struct {
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type MonthlySpendingFilter struct {
	UserID string
	From   time.Time
	To     time.Time
}

type CategorySpendingFilter struct {
	UserID   string
	FromDate time.Time
	ToDate   time.Time
}

type AnalyticsUsecase interface {
	GetMonthlySpending(ctx context.Context, filter MonthlySpendingFilter) ([]MonthlySpendingPoint, error)
	GetSpendingByCategory(ctx context.Context, filter CategorySpendingFilter) ([]CategorySpendingPoint, error)
}
