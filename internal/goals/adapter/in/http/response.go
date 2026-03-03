package http

import "time"

type GoalResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Currency     string    `json:"currency"`
	TargetAmount int64     `json:"target_amount"`
	TargetDate   *string   `json:"target_date,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CompletedAt  *string   `json:"completed_at,omitempty"`
}

type ListGoalsResponse struct {
	Goals    []GoalResponse `json:"goals"`
	PageSize int            `json:"page_size"`
	Offset   int            `json:"offset"`
}

type GoalContributionResponse struct {
	ID               string    `json:"id"`
	GoalID           string    `json:"goal_id"`
	UserID           string    `json:"user_id"`
	Amount           int64     `json:"amount"`
	ContributionDate string    `json:"contribution_date"`
	Note             string    `json:"note,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ListContributionsResponse struct {
	Contributions []GoalContributionResponse `json:"contributions"`
	PageSize      int                        `json:"page_size"`
	Offset        int                        `json:"offset"`
}

type GoalProgressResponse struct {
	GoalID             string  `json:"goal_id"`
	TargetAmount       int64   `json:"target_amount"`
	SavedAmount        int64   `json:"saved_amount"`
	RemainingAmount    int64   `json:"remaining_amount"`
	ProgressPercentage float64 `json:"progress_percentage"`
	IsCompleted        bool    `json:"is_completed"`
	LastUpdatedAt      string  `json:"last_updated_at"`
}
