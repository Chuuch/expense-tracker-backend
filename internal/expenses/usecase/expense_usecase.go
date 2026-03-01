package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/utils"
)

type ExpenseUsecase struct {
	expenseRepo interfaces.ExpenseRepository
}

func NewExpenseUsecase(expenseRepo interfaces.ExpenseRepository) *ExpenseUsecase {
	return &ExpenseUsecase{
		expenseRepo: expenseRepo,
	}
}

func (u *ExpenseUsecase) CreateExpense(
	ctx context.Context,
	userID string,
	amount int64,
	currency string,
	category domain.ExpenseCategory,
	description string,
	date time.Time,
) (*domain.Expense, error) {
	expense, err := domain.NewExpense(
		utils.GenerateULID(),
		userID,
		amount,
		currency,
		category,
		description,
		date,
	)
	if err != nil {
		return nil, fmt.Errorf("usecase.NewExpense: %w", err)
	}

	createdExpense, err := u.expenseRepo.CreateExpense(ctx, expense)
	if err != nil {
		return nil, fmt.Errorf("usecase.CreateExpense: %w", err)
	}
	return createdExpense, nil
}

func (u *ExpenseUsecase) GetExpenseByID(ctx context.Context, userID, expenseID string) (*domain.Expense, error) {
	expense, err := u.expenseRepo.GetExpenseByID(ctx, userID, expenseID)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetExpenseByID: %w", err)
	}
	if expense == nil {
		return nil, ErrExpenseNotFound
	}
	return expense, nil
}

func (u *ExpenseUsecase) ListExpenses(
	ctx context.Context,
	filter interfaces.ExpenseListFilter,
) ([]*domain.Expense, error) {
	if strings.TrimSpace(filter.UserID) == "" {
		return nil, ErrExpenseInvalidFilter
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
		return nil, ErrExpenseInvalidFilter
	}

	expenses, err := u.expenseRepo.ListExpenses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("usecase.ListExpenses: %w", err)
	}
	return expenses, nil
}

func (u *ExpenseUsecase) UpdateExpense(
	ctx context.Context,
	userID string,
	expenseID string,
	amount int64,
	currency string,
	category domain.ExpenseCategory,
	description string,
	date time.Time,
) (*domain.Expense, error) {
	existing, err := u.expenseRepo.GetExpenseByID(ctx, userID, expenseID)
	if err != nil {
		return nil, fmt.Errorf("usecase.UpdateExpense: %w", err)
	}
	if existing == nil {
		return nil, ErrExpenseNotFound
	}

	updated, err := domain.NewExpense(
		existing.ID,
		existing.UserID,
		amount,
		currency,
		category,
		description,
		date,
	)
	if err != nil {
		return nil, fmt.Errorf("usecase.NewExpense: %w", err)
	}

	updated.CreatedAt = existing.CreatedAt
	updated.UpdatedAt = time.Now()

	if err := u.expenseRepo.UpdateExpense(ctx, updated); err != nil {
		return nil, fmt.Errorf("usecase.UpdateExpense: %w", err)
	}
	return updated, nil
}

func (u *ExpenseUsecase) DeleteExpense(
	ctx context.Context,
	userID string,
	expenseID string,
) error {
	expense, err := u.expenseRepo.GetExpenseByID(ctx, userID, expenseID)
	if err != nil {
		return fmt.Errorf("usecase.GetExpenseByID: %w", err)
	}
	if expense == nil {
		return ErrExpenseNotFound
	}

	if err := u.expenseRepo.DeleteExpense(ctx, userID, expenseID); err != nil {
		if errors.Is(err, ErrExpenseNotFound) {
			return ErrExpenseNotFound
		}
		return fmt.Errorf("usecase.DeleteExpenes: %w", err)
	}
	return nil
}
