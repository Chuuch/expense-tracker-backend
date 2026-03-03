package postgres

import (
	"context"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

type AnalyticsRepository struct {
	q postgresdb.Querier
}

var _ interfaces.AnalyticsRepository = (*AnalyticsRepository)(nil)

func NewAnalyticsRepository(q postgresdb.Querier) *AnalyticsRepository {
	return &AnalyticsRepository{q: q}
}

func (r *AnalyticsRepository) GetMonthlySpending(
	ctx context.Context,
	filter interfaces.MonthlySpendingFilter,
) ([]interfaces.MonthlySpendingPoint, error) {
	toExclusive := filter.To.AddDate(0, 1, 0)
	rows, err := r.q.GetMonthlySpending(ctx, postgresdb.GetMonthlySpendingParams{
		UserID:   filter.UserID,
		FromDate: filter.From,
		ToDate:   toExclusive,
	})
	if err != nil {
		return nil, err
	}

	out := make([]interfaces.MonthlySpendingPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, interfaces.MonthlySpendingPoint{
			YearMonth: row.YearMonth,
			Amount:    row.Amount,
			Currency:  row.Currency,
		})
	}
	return out, nil
}

func (r *AnalyticsRepository) GetSpendingByCategory(
	ctx context.Context,
	filter interfaces.CategorySpendingFilter,
) ([]interfaces.CategorySpendingPoint, error) {
	from := filter.FromDate.Truncate(24 * time.Hour)
	to := filter.ToDate.Truncate(24 * time.Hour).Add(24 * time.Hour).Add(-time.Nanosecond)

	rows, err := r.q.GetSpendingByCategory(ctx, postgresdb.GetSpendingByCategoryParams{
		UserID:   filter.UserID,
		FromDate: from,
		ToDate:   to,
	})
	if err != nil {
		return nil, err
	}

	out := make([]interfaces.CategorySpendingPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, interfaces.CategorySpendingPoint{
			Category: row.Category,
			Amount:   row.Amount,
			Currency: row.Currency,
		})
	}
	return out, nil
}
