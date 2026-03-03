package http

import "time"

type MonthlySpendingQuery struct {
	FromDate string `query:"from_date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	ToDate   string `query:"to_date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

func (q MonthlySpendingQuery) ParseFromDate() (time.Time, error) {
	return time.Parse(time.RFC3339, q.FromDate)
}

func (q MonthlySpendingQuery) ParseToDate() (time.Time, error) {
	return time.Parse(time.RFC3339, q.ToDate)
}

type CategorySpendingQuery struct {
	FromDate string `query:"from_date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	ToDate   string `query:"to_date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

func (q CategorySpendingQuery) ParseFromDate() (time.Time, error) {
	return time.Parse(time.RFC3339, q.FromDate)
}

func (q CategorySpendingQuery) ParseToDate() (time.Time, error) {
	return time.Parse(time.RFC3339, q.ToDate)
}
