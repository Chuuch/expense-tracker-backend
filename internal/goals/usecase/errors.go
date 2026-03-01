package usecase

import "errors"

var (
	ErrGoalNotFound         = errors.New("goal not found")
	ErrGoalForbidden        = errors.New("goal forbidden")
	ErrGoalInvalidFilter    = errors.New("goal invalid filter")
	ErrGoalInvalidStatus    = errors.New("goal invalid status")
	ErrGoalAlreadyCompleted = errors.New("goal already completed")

	ErrGoalContributionForbidden = errors.New("goal contribution forbidden")
	ErrGoalContributionInvalid   = errors.New("goal contribution invalid")

	ErrGoalTargetReached = errors.New("goal target already reached")
)
