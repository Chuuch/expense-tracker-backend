package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrGoalContributionIDRequired     = errors.New("goal contribution id is required")
	ErrGoalContributionGoalIDRequired = errors.New("goal contribution goal id is required")
	ErrGoalContributionUserIDRequired = errors.New("goal contribution user id is required")
	ErrGoalContributionAmountInvalid  = errors.New("goal contribution amount must be greater than 0")
	ErrGoalContributionDateRequired   = errors.New("goal contribution date is required")
)

type GoalContribution struct {
	ID               string
	GoalID           string
	UserID           string
	Amount           int64
	ContributionDate time.Time
	Note             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewGoalContribution(
	id string,
	goalID string,
	userID string,
	amount int64,
	contributionDate time.Time,
	note string,
) (*GoalContribution, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrGoalContributionIDRequired
	}
	if strings.TrimSpace(goalID) == "" {
		return nil, ErrGoalContributionGoalIDRequired
	}
	if strings.TrimSpace(userID) == "" {
		return nil, ErrGoalContributionUserIDRequired
	}
	if amount <= 0 {
		return nil, ErrGoalContributionAmountInvalid
	}
	if contributionDate.IsZero() {
		return nil, ErrGoalContributionDateRequired
	}

	now := time.Now()
	return &GoalContribution{
		ID:               strings.TrimSpace(id),
		GoalID:           strings.TrimSpace(goalID),
		UserID:           strings.TrimSpace(userID),
		Amount:           amount,
		ContributionDate: contributionDate,
		Note:             strings.TrimSpace(note),
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}
