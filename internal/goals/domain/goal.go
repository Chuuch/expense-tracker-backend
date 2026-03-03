package domain

import (
	"errors"
	"strings"
	"time"
)

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusInactive  GoalStatus = "inactive"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusCancelled GoalStatus = "cancelled"
)

var (
	ErrGoalIDRequired          = errors.New("goal id is required")
	ErrGoalUserIDRequired      = errors.New("goal user id is required")
	ErrGoalNameRequired        = errors.New("goal name is required")
	ErrGoalCurrencyIsRequired  = errors.New("goal currency is required")
	ErrGoalTargetAmountInvalid = errors.New("goal target amount must be greater than 0")
	ErrGoalStatusInvalid       = errors.New("goal status is invalid")
	ErrGoalStatusTransition    = errors.New("goal status transition is not allowed")
)

type Goal struct {
	ID           string
	UserID       string
	Name         string
	Currency     string
	TargetAmount int64
	TargetDate   *time.Time
	Status       GoalStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

func NewGoal(
	id string,
	userID string,
	name string,
	currency string,
	targetAmount int64,
	targetDate *time.Time,
) (*Goal, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGoalIDRequired
	}
	if strings.TrimSpace(userID) == "" {
		return nil, ErrGoalUserIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrGoalNameRequired
	}
	if strings.TrimSpace(currency) == "" {
		return nil, ErrGoalCurrencyIsRequired
	}
	if targetAmount <= 0 {
		return nil, ErrGoalTargetAmountInvalid
	}

	now := time.Now()
	return &Goal{
		ID:           strings.TrimSpace(id),
		UserID:       strings.TrimSpace(userID),
		Name:         strings.TrimSpace(name),
		Currency:     NormalizeCurrency(currency),
		TargetAmount: targetAmount,
		TargetDate:   targetDate,
		Status:       GoalStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		CompletedAt:  nil,
	}, nil
}

func NormalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func IsValidGoalStatus(status GoalStatus) bool {
	switch status {
	case GoalStatusActive, GoalStatusInactive, GoalStatusCompleted, GoalStatusCancelled:
		return true
	default:
		return false
	}
}

func (g *Goal) CanTransitionTo(next GoalStatus) bool {
	if g == nil {
		return false
	}
	if !IsValidGoalStatus(next) {
		return false
	}

	switch g.Status {
	case GoalStatusActive:
		return next == GoalStatusInactive || next == GoalStatusCompleted || next == GoalStatusCancelled
	case GoalStatusInactive:
		return next == GoalStatusActive || next == GoalStatusCancelled
	case GoalStatusCompleted:
		return false
	default:
		return false
	}
}

func (g *Goal) TransitionTo(next GoalStatus, now time.Time) error {
	if g == nil {
		return ErrGoalStatusInvalid
	}
	if !IsValidGoalStatus(next) {
		return ErrGoalStatusInvalid
	}

	g.Status = next
	g.UpdatedAt = now

	if next == GoalStatusCompleted {
		g.CompletedAt = &now
	}
	return nil
}
