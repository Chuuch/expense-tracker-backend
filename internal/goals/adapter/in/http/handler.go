package http

import (
	"net/http"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	goalerrors "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/in/http/errors"
	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	"github.com/labstack/echo/v5"
)

type GoalHandler struct {
	usecase interfaces.GoalUsecase
}

func NewGoalHandler(u interfaces.GoalUsecase) *GoalHandler {
	return &GoalHandler{
		usecase: u,
	}
}

func (h *GoalHandler) CreateGoal(c *echo.Context) error {
	var req CreateGoalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Validation failed"})
	}

	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	targetDate, err := req.ParseTargetDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid target date"})
	}

	goal, err := h.usecase.CreateGoal(
		c.Request().Context(),
		userID,
		req.Name,
		req.Currency,
		req.TargetAmount,
		targetDate,
	)
	if err != nil {
		status, resp := goalerrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.JSON(http.StatusCreated, mapGoalToResponse(goal))
}

func (h *GoalHandler) GetGoalByID(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	goalID := c.Param("id")
	goal, err := h.usecase.GetGoalByID(c.Request().Context(), userID, goalID)
	if err != nil {
		status, resp := goalerrors.Map(err)
		return c.JSON(status, resp)
	}
	return c.JSON(http.StatusOK, mapGoalToResponse(goal))
}

func (h *GoalHandler) ListGoals(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	var query ListGoalsQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid query params"})
	}
	if err := c.Validate(&query); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Validation failed"})
	}

	status, err := query.ParseStatus()
	if err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid status"})
	}

	filter := interfaces.GoalListFilter{
		UserID:   userID,
		Status:   status,
		PageSize: query.PageSize,
		Offset:   query.Offset,
	}

	goals, err := h.usecase.ListGoals(c.Request().Context(), filter)
	if err != nil {
		code, resp := goalerrors.Map(err)
		return c.JSON(code, resp)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	return c.JSON(http.StatusOK, mapGoalsToResponse(goals, pageSize, offset))
}

func (h *GoalHandler) UpdateGoal(c *echo.Context) error {
	var req UpdateGoalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Validation failed"})
	}

	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	targetDate, err := req.ParseTargetDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid target_date format"})
	}

	status := domain.GoalStatus(req.Status)
	goalID := c.Param("id")

	goal, err := h.usecase.UpdateGoal(
		c.Request().Context(),
		userID,
		goalID,
		req.Name,
		req.TargetAmount,
		targetDate,
		status,
	)

	if err != nil {
		code, resp := goalerrors.Map(err)
		return c.JSON(code, resp)
	}
	return c.JSON(http.StatusOK, mapGoalToResponse(goal))
}

func (h *GoalHandler) DeleteGoal(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	goalID := c.Param("id")
	if err := h.usecase.DeleteGoal(c.Request().Context(), userID, goalID); err != nil {
		code, resp := goalerrors.Map(err)
		return c.JSON(code, resp)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *GoalHandler) AddContribution(c *echo.Context) error {
	var req AddContributionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Validation failed"})
	}

	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	contributionDate, err := req.ParseContributionDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid contribution date format"})
	}

	goalID := c.Param("id")
	contribution, err := h.usecase.AddContribution(
		c.Request().Context(),
		userID,
		goalID,
		req.Amount,
		contributionDate,
		req.Note,
	)

	if err != nil {
		code, resp := goalerrors.Map(err)
		return c.JSON(code, resp)
	}

	return c.JSON(http.StatusCreated, mapContributionToResponse(contribution))
}

func (h *GoalHandler) ListContributions(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	var query ListContributionsQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Invalid query params"})
	}

	if err := c.Validate(&query); err != nil {
		return c.JSON(http.StatusBadRequest, goalerrors.Response{Error: "Validation failed"})
	}

	goalID := c.Param("id")
	contributions, err := h.usecase.ListContributions(
		c.Request().Context(),
		userID,
		goalID,
		query.PageSize,
		query.Offset,
	)
	if err != nil {
		code, resp := goalerrors.Map(err)
		return c.JSON(code, resp)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	return c.JSON(http.StatusOK, mapContributionsToResponse(contributions, pageSize, offset))
}

func (h *GoalHandler) GetGoalProgress(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, goalerrors.Response{Error: "Unauthorized"})
	}

	goalID := c.Param("id")
	progress, err := h.usecase.GetGoalProgress(c.Request().Context(), userID, goalID)
	if err != nil {
		code, resp := goalerrors.Map(err)
		return c.JSON(code, resp)
	}
	return c.JSON(http.StatusOK, mapProgressToResponse(progress))
}
