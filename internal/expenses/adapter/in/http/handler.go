package http

import (
	"net/http"
	"time"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	expenseerrors "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http/errors"
	"github.com/chuuch/expense-tracker-backend/internal/expenses/usecase/interfaces"
	"github.com/labstack/echo/v5"
)

type ExpenseHandler struct {
	usecase interfaces.ExpenseUsecase
}

func NewExpenseHandler(usecase interfaces.ExpenseUsecase) *ExpenseHandler {
	return &ExpenseHandler{
		usecase: usecase,
	}
}

func (h *ExpenseHandler) CreateExpense(c *echo.Context) error {
	var req CreateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Validation failed"})
	}

	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, expenseerrors.Response{Error: "Unauthorized"})
	}

	category, err := parseCategory(req.Category)
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid category"})
	}

	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid date format"})
	}

	expense, err := h.usecase.CreateExpense(
		c.Request().Context(),
		userID,
		req.Amount,
		req.Currency,
		category,
		req.Description,
		date,
	)

	if err != nil {
		code, payload := expenseerrors.Map(err)
		return c.JSON(code, payload)
	}

	return c.JSON(http.StatusCreated, mapExpenseToResponse(expense))
}

func (h *ExpenseHandler) GetExpenseByID(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, expenseerrors.Response{Error: "Unauthorized"})
	}

	expenseID := c.Param("id")
	expense, err := h.usecase.GetExpenseByID(c.Request().Context(), userID, expenseID)
	if err != nil {
		code, payload := expenseerrors.Map(err)
		return c.JSON(code, payload)
	}
	return c.JSON(http.StatusOK, mapExpenseToResponse(expense))
}

func (h *ExpenseHandler) ListExpenses(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, expenseerrors.Response{Error: "Unauthorized"})
	}

	var query ListExpensesQuery
	if err := c.Bind(&query); err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid request body"})
	}

	if err := c.Validate(&query); err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Validation failed"})
	}

	category, err := parseCategory(query.Category)
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid category"})
	}

	fromDate, err := query.ParseFromDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid from_date format"})
	}

	toDate, err := query.ParseToDate()
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid to_date format"})
	}

	filter := interfaces.ExpenseListFilter{
		UserID:   userID,
		FromDate: fromDate,
		ToDate:   toDate,
		Limit:    query.Limit,
		Offset:   query.Offset,
	}
	if category != "" {
		filter.Category = &category
	}

	expenses, err := h.usecase.ListExpenses(c.Request().Context(), filter)
	if err != nil {
		code, payload := expenseerrors.Map(err)
		return c.JSON(code, payload)
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	return c.JSON(http.StatusOK, mapExpensesToResponse(expenses, limit, offset))
}

func (h *ExpenseHandler) UpdateExpense(c *echo.Context) error {
	var req UpdateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Validation failed"})
	}

	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, expenseerrors.Response{Error: "Unauthorized"})
	}

	expenseID := c.Param("id")

	category, err := parseCategory(req.Category)
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid category"})
	}

	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return c.JSON(http.StatusBadRequest, expenseerrors.Response{Error: "Invalid date format"})
	}

	expense, err := h.usecase.UpdateExpense(
		c.Request().Context(),
		userID,
		expenseID,
		req.Amount,
		req.Currency,
		category,
		req.Description,
		date,
	)
	if err != nil {
		code, payload := expenseerrors.Map(err)
		return c.JSON(code, payload)
	}
	return c.JSON(http.StatusOK, mapExpenseToResponse(expense))
}

func (h *ExpenseHandler) DeleteExpense(c *echo.Context) error {
	userID, ok := authhttp.GetUserIDFromContext(c.Request().Context())
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, expenseerrors.Response{Error: "Unauthorized"})
	}

	expenseID := c.Param("id")
	if err := h.usecase.DeleteExpense(c.Request().Context(), userID, expenseID); err != nil {
		code, payload := expenseerrors.Map(err)
		return c.JSON(code, payload)
	}

	return c.NoContent(http.StatusNoContent)
}
