package errors

import (
	"errors"
	"net/http"

	goaldomain "github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	goalusecase "github.com/chuuch/expense-tracker-backend/internal/goals/usecase"
)

type Response struct {
	Error string `json:"error"`
}

func Map(err error) (int, Response) {
	switch {
	case errors.Is(err, goalusecase.ErrGoalNotFound):
		return http.StatusNotFound, Response{Error: "Goal not found"}

	case errors.Is(err, goalusecase.ErrGoalForbidden):
		return http.StatusForbidden, Response{Error: "Forbidden"}

	case errors.Is(err, goalusecase.ErrGoalInvalidFilter),
		errors.Is(err, goalusecase.ErrGoalInvalidStatus),
		errors.Is(err, goalusecase.ErrGoalContributionInvalid),
		errors.Is(err, goaldomain.ErrGoalNameRequired),
		errors.Is(err, goaldomain.ErrGoalCurrencyIsRequired),
		errors.Is(err, goaldomain.ErrGoalTargetAmountInvalid),
		errors.Is(err, goaldomain.ErrGoalStatusInvalid),
		errors.Is(err, goaldomain.ErrGoalContributionAmountInvalid),
		errors.Is(err, goaldomain.ErrGoalContributionDateRequired):
		return http.StatusBadRequest, Response{Error: "Invalid request"}

	case errors.Is(err, goalusecase.ErrGoalAlreadyCompleted),
		errors.Is(err, goalusecase.ErrGoalTargetReached):
		return http.StatusConflict, Response{Error: "Goal already completed"}

	default:
		return http.StatusInternalServerError, Response{Error: "Internal server error"}
	}
}
