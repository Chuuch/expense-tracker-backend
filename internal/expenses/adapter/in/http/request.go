package http

import "time"

type CreateExpenseRequest struct {
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Currency    string `json:"currency" validate:"required,len=3"`
	Category    string `json:"category" validate:"required"`
	Description string `json:"description"`
	Date        string `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type UpdateExpenseRequest struct {
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Currency    string `json:"currency" validate:"required,len=3"`
	Category    string `json:"category" validate:"required"`
	Description string `json:"description"`
	Date        string `json:"date" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

type ListExpensesQuery struct {
	Category string `query:"category"`
	FromDate string `query:"from_date" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	ToDate   string `query:"to_date" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Limit    int    `query:"limit" validate:"omitempty,gte=1,lte=100"`
	Offset   int    `query:"offset" validate:"omitempty,gte=0"`
}

func (q ListExpensesQuery) ParseFromDate() (*time.Time, error) {
	if q.FromDate == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, q.FromDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (q ListExpensesQuery) ParseToDate() (*time.Time, error) {
	if q.ToDate == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, q.ToDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
