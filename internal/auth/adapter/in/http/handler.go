package http

import (
	"net/http"

	httperrors "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http/errors"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	usecase        interfaces.UserUsecase
	tokenUsecase   interfaces.TokenUsecase
	refreshUsecase interfaces.RefreshTokenUsecase
	cfg            *config.Config
}

func NewUserHandler(usecase interfaces.UserUsecase, tokenUsecase interfaces.TokenUsecase, cfg *config.Config) *UserHandler {
	return &UserHandler{
		usecase:      usecase,
		tokenUsecase: tokenUsecase,
		cfg:          cfg,
	}
}

func NewUserHandlerWithRefresh(
	usecase interfaces.UserUsecase,
	tokenUsecase interfaces.TokenUsecase,
	refreshUsecase interfaces.RefreshTokenUsecase,
	cfg *config.Config,
) *UserHandler {
	return &UserHandler{
		usecase:        usecase,
		tokenUsecase:   tokenUsecase,
		refreshUsecase: refreshUsecase,
		cfg:            cfg,
	}
}

func (h *UserHandler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Validation failed"})
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
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.JSON(http.StatusCreated, mapUserToResponse(user))
}

func (h *UserHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Invalid credentials format"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Validation failed"})
	}

	user, err := h.usecase.Login(
		c.Request().Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	if h.refreshUsecase != nil {
		pair, err := h.refreshUsecase.IssueTokenPair(c.Request().Context(), user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to issue token pair"})
		}
		return c.JSON(http.StatusOK, LoginResponse{
			User:         mapUserToResponse(user),
			Token:        pair.AccessToken,
			RefreshToken: pair.RefreshToken,
		})
	}

	token, err := h.tokenUsecase.GenerateToken(user, h.cfg.Auth.AccessTokenTTL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, LoginResponse{
		User:  mapUserToResponse(user),
		Token: token,
	})
}

func (h *UserHandler) Refresh(c *echo.Context) error {
	if h.refreshUsecase == nil {
		return c.JSON(http.StatusNotImplemented, httperrors.Response{Error: "Refresh token flow not configured"})
	}

	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Validation failed"})
	}

	pair, err := h.refreshUsecase.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.JSON(http.StatusOK, RefreshResponse{
		Token:        pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *UserHandler) Logout(c *echo.Context) error {
	if h.refreshUsecase == nil {
		return c.JSON(http.StatusNotImplemented, httperrors.Response{Error: "Refresh token flow not configured"})
	}

	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Validation failed"})
	}

	if err := h.refreshUsecase.Revoke(c.Request().Context(), req.RefreshToken); err != nil {
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	user, err := h.usecase.GetByID(c.Request().Context(), id)
	if err != nil {
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.JSON(http.StatusOK, mapUserToResponse(user))
}

func (h *UserHandler) UpdateUser(c *echo.Context) error {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Validation failed"})
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
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.JSON(http.StatusOK, mapUserToResponse(user))
}

func (h *UserHandler) DeleteUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.usecase.DeleteUser(c.Request().Context(), id); err != nil {
		status, resp := httperrors.Map(err)
		return c.JSON(status, resp)
	}

	return c.NoContent(http.StatusNoContent)
}
