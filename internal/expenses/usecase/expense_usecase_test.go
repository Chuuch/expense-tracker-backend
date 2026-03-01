package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/expenses/domain"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces/mocks"
	"go.uber.org/mock/gomock"
)

func TestExpenseUsecase_CreateExpense_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()
	date := time.Now()

	repo.EXPECT().
		CreateExpense(ctx, gomock.AssignableToTypeOf(&domain.Expense{})).
		DoAndReturn(func(_ context.Context, e *domain.Expense) (*domain.Expense, error) {
			if e.ID == "" {
				t.Fatalf("expected generated expense ID")
			}
			if e.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", e.UserID)
			}
			if e.Amount != 12345 {
				t.Fatalf("expected amount 12345, got %d", e.Amount)
			}
			if e.Currency != "USD" {
				t.Fatalf("expected normalized currency USD, got %s", e.Currency)
			}
			if e.Category != domain.CategoryHousing {
				t.Fatalf("expected category housing, got %s", e.Category)
			}
			return e, nil
		}).
		Times(1)

	got, err := uc.CreateExpense(
		ctx,
		"u1",
		12345,
		"usd",
		domain.CategoryHousing,
		"Rent",
		date,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected created expense, got nil")
	}
}

func TestExpenseUsecase_GetExpenseByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().
		GetExpenseByID(ctx, "u1", "e_missing").
		Return(nil, nil).
		Times(1)

	_, err := uc.GetExpenseByID(ctx, "u1", "e_missing")
	if !errors.Is(err, usecase.ErrExpenseNotFound) {
		t.Fatalf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestExpenseUsecase_ListExpenses_InvalidFilterMissingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	_, err := uc.ListExpenses(ctx, interfaces.ExpenseListFilter{
		UserID: "",
		Limit:  20,
		Offset: 0,
	})
	if !errors.Is(err, usecase.ErrExpenseInvalidFilter) {
		t.Fatalf("expected ErrExpenseInvalidFilter, got %v", err)
	}
}

func TestExpenseUsecase_ListExpenses_AppliesDefaultsAndCapsLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().
		ListExpenses(ctx, gomock.AssignableToTypeOf(interfaces.ExpenseListFilter{})).
		DoAndReturn(func(_ context.Context, f interfaces.ExpenseListFilter) ([]*domain.Expense, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if f.Limit != 100 {
				t.Fatalf("expected capped limit 100, got %d", f.Limit)
			}
			if f.Offset != 0 {
				t.Fatalf("expected normalized offset 0, got %d", f.Offset)
			}
			return []*domain.Expense{}, nil
		}).
		Times(1)

	_, err := uc.ListExpenses(ctx, interfaces.ExpenseListFilter{
		UserID: "u1",
		Limit:  999,
		Offset: -5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpenseUsecase_ListExpenses_InvalidDateRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	from := time.Now()
	to := from.Add(-1 * time.Hour)

	_, err := uc.ListExpenses(ctx, interfaces.ExpenseListFilter{
		UserID:   "u1",
		FromDate: &from,
		ToDate:   &to,
		Limit:    20,
	})
	if !errors.Is(err, usecase.ErrExpenseInvalidFilter) {
		t.Fatalf("expected ErrExpenseInvalidFilter, got %v", err)
	}
}

func TestExpenseUsecase_UpdateExpense_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().
		GetExpenseByID(ctx, "u1", "e_missing").
		Return(nil, nil).
		Times(1)

	_, err := uc.UpdateExpense(
		ctx,
		"u1",
		"e_missing",
		1500,
		"USD",
		domain.CategoryFood,
		"Lunch",
		time.Now(),
	)
	if !errors.Is(err, usecase.ErrExpenseNotFound) {
		t.Fatalf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestExpenseUsecase_UpdateExpense_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	existing := &domain.Expense{
		ID:          "e1",
		UserID:      "u1",
		Amount:      1000,
		Currency:    "USD",
		Category:    domain.CategoryFood,
		Description: "old",
		Date:        time.Now().Add(-24 * time.Hour),
		CreatedAt:   time.Now().Add(-48 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}

	repo.EXPECT().
		GetExpenseByID(ctx, "u1", "e1").
		Return(existing, nil).
		Times(1)

	repo.EXPECT().
		UpdateExpense(ctx, gomock.AssignableToTypeOf(&domain.Expense{})).
		DoAndReturn(func(_ context.Context, e *domain.Expense) error {
			if e.ID != "e1" || e.UserID != "u1" {
				t.Fatalf("expected immutable id/user to be preserved, got id=%s user=%s", e.ID, e.UserID)
			}
			if !e.CreatedAt.Equal(existing.CreatedAt) {
				t.Fatalf("expected CreatedAt preserved")
			}
			if e.Amount != 2000 {
				t.Fatalf("expected amount 2000, got %d", e.Amount)
			}
			if e.Currency != "USD" {
				t.Fatalf("expected normalized currency USD, got %s", e.Currency)
			}
			if e.Category != domain.CategoryHousing {
				t.Fatalf("expected category housing, got %s", e.Category)
			}
			return nil
		}).
		Times(1)

	got, err := uc.UpdateExpense(
		ctx,
		"u1",
		"e1",
		2000,
		"usd",
		domain.CategoryHousing,
		"Rent",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected updated expense, got nil")
	}
}

func TestExpenseUsecase_DeleteExpense_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().
		GetExpenseByID(ctx, "u1", "e_missing").
		Return(nil, nil).
		Times(1)

	err := uc.DeleteExpense(ctx, "u1", "e_missing")
	if !errors.Is(err, usecase.ErrExpenseNotFound) {
		t.Fatalf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestExpenseUsecase_DeleteExpense_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	repo.EXPECT().
		GetExpenseByID(ctx, "u1", "e1").
		Return(&domain.Expense{ID: "e1", UserID: "u1"}, nil).
		Times(1)

	repo.EXPECT().
		DeleteExpense(ctx, "u1", "e1").
		Return(nil).
		Times(1)

	if err := uc.DeleteExpense(ctx, "u1", "e1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpenseUsecase_CreateExpense_InvalidAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockExpenseRepository(ctrl)
	uc := usecase.NewExpenseUsecase(repo)
	ctx := context.Background()

	// CreateExpense should fail at domain validation and never hit repo.
	repo.EXPECT().
		CreateExpense(gomock.Any(), gomock.Any()).
		Times(0)

	_, err := uc.CreateExpense(
		ctx,
		"u1",
		0, // invalid
		"USD",
		domain.CategoryFood,
		"Lunch",
		time.Now(),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
