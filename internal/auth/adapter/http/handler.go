package http

import (
	"net/http"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/ports"
	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	usecase      ports.UserUsecase
	tokenUsecase ports.TokenUsecase
}

func NewUserHandler(usecase ports.UserUsecase, tokenUsecase ports.TokenUsecase) *UserHandler {
	return &UserHandler{
		usecase:      usecase,
		tokenUsecase: tokenUsecase,
	}
}

func (h *UserHandler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	user, err := h.usecase.Register(
		c.Request().Context(),
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		req.Phone,
		req.Address,
		req.City,
		req.State,
		req.Zip,
		req.Country,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusBadRequest, mapUserToResponse(user))
}

func (h *UserHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid credentils format"})
	}

	user, err := h.usecase.Login(
		c.Request().Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	}

	token, err := h.tokenUsecase.GenerateToken(user, 24*time.Hour)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		User:  mapUserToResponse(user),
		Token: token,
	})
}

func (h *UserHandler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	user, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
	}

	return c.JSON(http.StatusOK, mapUserToResponse(user))
}

func (h *UserHandler) UpdateUser(c *echo.Context) error {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	user, err := h.usecase.UpdateUser(
		c.Request().Context(),
		id,
		req.FirstName,
		req.LastName,
		req.Phone,
		req.Address,
		req.City,
		req.State,
		req.Zip,
		req.Country,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, mapUserToResponse(user))
}

func (h *UserHandler) DeleteUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.usecase.DeleteUser(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func mapUserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.Profile.FirstName,
		LastName:     u.Profile.LastName,
		Phone:        u.Profile.Phone,
		Address:      u.Profile.Address,
		City:         u.Profile.City,
		State:        u.Profile.State,
		Zip:          u.Profile.Zip,
		Country:      u.Profile.Country,
		Role:         string(u.Role),
		Status:       string(u.Status),
		IsMFAEnabled: u.IsMFAEnabled,
		CreatedAt:    u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
