package http

import (
	"net/http"

	analyticserrors "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/in/http/errors"
	"github.com/chuuch/expense-tracker-backend/internal/analytics/usecase/interfaces"
	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/labstack/echo/v5"
)

type AnalyticsHandler struct {
	u interfaces.AnalyticsUsecase
}

func NewAnalyticsHandler(u interfaces.AnalyticsUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{u: u}
}

func (h *AnalyticsHandler) GetMonthlySpending(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, analyticserrors.Response{Error: "Unauthorized"})
	}

	var query MonthlySpendingQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Invaild query params"})
	}

	if err := c.Validate(&query); err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Validation failed"})
	}

	from, err := query.ParseFromDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Invalid from_date format"})
	}

	to, err := query.ParseToDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Invalid to_date format"})
	}

	filter := interfaces.MonthlySpendingFilter{
		UserID: userID,
		From:   from,
		To:     to,
	}

	points, err := h.u.GetMonthlySpending(c.Request().Context(), filter)
	if err != nil {
		status, response := analyticserrors.Map(err)
		return c.JSON(status, response)
	}

	return c.JSON(http.StatusOK, mapMonthlySpendingToResponse(points))
}

func (h *AnalyticsHandler) GetSpendingByCategory(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, analyticserrors.Response{Error: "Unauthorized"})
	}

	var query CategorySpendingQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Invalid query params"})
	}

	if err := c.Validate(&query); err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Validation failed"})
	}

	from, err := query.ParseFromDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Invalid from_date format"})
	}

	to, err := query.ParseToDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, analyticserrors.Response{Error: "Invalid to_date format"})
	}

	filter := interfaces.CategorySpendingFilter{
		UserID:   userID,
		FromDate: from,
		ToDate:   to,
	}

	points, err := h.u.GetSpendingByCategory(c.Request().Context(), filter)
	if err != nil {
		status, response := analyticserrors.Map(err)
		return c.JSON(status, response)
	}

	return c.JSON(http.StatusOK, mapCategorySpendingToResponse(points))
}

func mapMonthlySpendingToResponse(points []interfaces.MonthlySpendingPoint) MonthlySpendingResponse {
	items := make([]MonthlySpendingPointResponse, 0, len(points))
	for _, p := range points {
		items = append(items, MonthlySpendingPointResponse{
			YearMonth: p.YearMonth,
			Amount:    p.Amount,
			Currency:  p.Currency,
		})
	}
	return MonthlySpendingResponse{
		Points: items,
	}
}

func mapCategorySpendingToResponse(points []interfaces.CategorySpendingPoint) CategorySpendingResponse {
	items := make([]CategorySpendingPointResponse, 0, len(points))
	for _, p := range points {
		items = append(items, CategorySpendingPointResponse{
			Category: p.Category,
			Amount:   p.Amount,
			Currency: p.Currency,
		})
	}
	return CategorySpendingResponse{
		Points: items,
	}
}
