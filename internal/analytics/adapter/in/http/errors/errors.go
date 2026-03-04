package errors

import (
	"errors"
	"net/http"

	analyticsusecase "github.com/chuuch/expense-tracker-backend/internal/analytics/usecase"
)

type Response struct {
	Error string `json:"error"`
}

func Map(err error) (int, Response) {
	switch {
	case errors.Is(err, analyticsusecase.ErrAnalyticsInvalidFilter):
		return http.StatusBadRequest, Response{Error: "Invalid filters"}

	case errors.Is(err, analyticsusecase.ErrAnalyticsForbidden):
		return http.StatusForbidden, Response{Error: "Forbidden"}

	case errors.Is(err, analyticsusecase.ErrAnalyticsNoData):
		return http.StatusNotFound, Response{Error: "No analytics data"}

	default:
		return http.StatusInternalServerError, Response{Error: "Internal server error"}
	}
}
