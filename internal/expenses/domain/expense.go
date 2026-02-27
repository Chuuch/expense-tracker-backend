package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Expense struct {
	ID          string
	UserID      string
	Amount      decimal.Decimal
	Currency    string
	Category    string
	Description string
	Date        time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewExpense(
	id string,
	userID string,
	amount decimal.Decimal,
	currency string,
	category string,
	description string,
	date time.Time,
) *Expense {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	if date.IsZero() {
		date = time.Now()
	}

	return &Expense{
		ID:          id,
		UserID:      userID,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		Description: description,
		Date:        date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
