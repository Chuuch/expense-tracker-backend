package http

import (
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
)

type CreateGoalRequest struct {
	Name         string  `json:"name" validate:"required,min=3,max=120"`
	Currency     string  `json:"currency" validate:"required,len=3"`
	TargetAmount int64   `json:"target_amount" validate:"required,gt=0"`
	TargetDate   *string `json:"target_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateGoalRequest struct {
	Name         string  `json:"nae" validate:"required,min=3,max=120"`
	TargetAmount int64   `json:"target_amount" validate:"required,gt=0"`
	TargetDate   *string `json:"target_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Status       string  `json:"status" validate:"required,oneof=active completed cancelled inactive"`
}

type AddContributionRequest struct {
	Amount           int64  `json:"amount" validate:"required,gt=0"`
	ContributionDate string `json:"contribution_date" validate:"required,datetime=2006-01-02"`
	Note             string `json:"note" validate:"omitempty,max=255"`
}

type ListGoalsQuery struct {
	Status   string `query:"status" validate:"omitempty,oneof=active completed cancelled inactive"`
	PageSize int    `query:"page_size" validate:"omitempty,gte=1,lte=100"`
	Offset   int    `query:"offset" validate:"omitempty,gte=0"`
}

type ListContributionsQuery struct {
	PageSize int `query:"page_size" validate:"omitempty,gte=1,lte=100"`
	Offset   int `query:"offset" validate:"omitempty,gte=0"`
}

func (r *CreateGoalRequest) ParseTargetDate() (*time.Time, error) {
	if r.TargetDate == nil || *r.TargetDate == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *r.TargetDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *UpdateGoalRequest) ParseTargetDate() (*time.Time, error) {
	if r.TargetDate == nil || *r.TargetDate == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *r.TargetDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *AddContributionRequest) ParseContributionDate() (time.Time, error) {
	return time.Parse("2006-01-02", r.ContributionDate)
}

func (q *ListGoalsQuery) ParseStatus() (*domain.GoalStatus, error) {
	if q.Status == "" {
		return nil, nil
	}

	status := domain.GoalStatus(q.Status)
	if !domain.IsValidGoalStatus(status) {
		return nil, domain.ErrGoalStatusInvalid
	}
	return &status, nil
}
