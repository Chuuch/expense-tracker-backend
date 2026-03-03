package http

type MonthlySpendingPointResponse struct {
	YearMonth string `json:"year_month"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type MonthlySpendingResponse struct {
	Points []MonthlySpendingPointResponse `json:"points"`
}

type CategorySpendingPointResponse struct {
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type CategorySpendingResponse struct {
	Points []CategorySpendingPointResponse `json:"points"`
}
