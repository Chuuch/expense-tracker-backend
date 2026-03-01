package domain

import (
	"errors"
	"strings"
	"time"
)

type ExpenseCategory string

const (
	CategoryFood          ExpenseCategory = "food"
	CategoryTransport     ExpenseCategory = "transport"
	CategoryHousing       ExpenseCategory = "housing"
	CategoryUtilities     ExpenseCategory = "utilities"
	CategoryHealthcare    ExpenseCategory = "healthcare"
	CategoryEntertainment ExpenseCategory = "entertainment"
	CategoryOther         ExpenseCategory = "other"
)

var (
	ErrExpenseIDRequired     = errors.New("expense id is required")
	ErrExpenseUserIDRequired = errors.New("expense user id is required")
	ErrExpenseAmountInvalid  = errors.New("expense amount must be greater than 0")
	ErrExpenseCurrencyEmpty  = errors.New("expense currenty is required")
	ErrExpenseDateRequired   = errors.New("expense date is required")
)

type Expense struct {
	ID          string
	UserID      string
	Amount      int64
	Currency    string
	Category    ExpenseCategory
	Description string
	Date        time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewExpense(
	id string,
	userID string,
	amount int64,
	currency string,
	category ExpenseCategory,
	description string,
	date time.Time,
) (*Expense, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrExpenseIDRequired
	}
	if strings.TrimSpace(userID) == "" {
		return nil, ErrExpenseIDRequired
	}
	if amount <= 0 {
		return nil, ErrExpenseAmountInvalid
	}
	if strings.TrimSpace(currency) == "" {
		return nil, ErrExpenseCurrencyEmpty
	}
	if date.IsZero() {
		return nil, ErrExpenseDateRequired
	}

	now := time.Now()
	return &Expense{
		ID:          id,
		UserID:      userID,
		Amount:      amount,
		Currency:    strings.ToUpper(strings.TrimSpace(currency)),
		Category:    category,
		Description: strings.TrimSpace(description),
		Date:        date,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
