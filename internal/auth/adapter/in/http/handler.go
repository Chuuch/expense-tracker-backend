package http

import (
	"net/http"
	"net/url"

	httperrors "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http/errors"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type UserHandler struct {
	usecase              interfaces.UserUsecase
	tokenUsecase         interfaces.TokenUsecase
	refreshUsecase       interfaces.RefreshTokenUsecase
	accessTokenBlacklist interfaces.AccessTokenBlacklist
	cfg                  *config.Config
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
	accessTokenBlacklist ...interfaces.AccessTokenBlacklist,
) *UserHandler {
	var blacklist interfaces.AccessTokenBlacklist
	if len(accessTokenBlacklist) > 0 {
		blacklist = accessTokenBlacklist[0]
	}

	return &UserHandler{
		usecase:              usecase,
		tokenUsecase:         tokenUsecase,
		refreshUsecase:       refreshUsecase,
		accessTokenBlacklist: blacklist,
		cfg:                  cfg,
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
		req.Username,
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

	if h.accessTokenBlacklist != nil {
		if accessToken, ok := extractBearerToken(c.Request().Header.Get(AuthHeader)); ok {
			if err := h.accessTokenBlacklist.Add(c.Request().Context(), accessToken, h.cfg.Auth.AccessTokenTTL); err != nil {
				return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to invalidate access token"})
			}
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) GoogleOAuthStart(c *echo.Context) error {
	redirectURI := c.QueryParam("redirect_uri")
	provider, err := goth.GetProvider("google")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Google provider not configured"})
	}

	state := utils.GenerateULID()

	sess, err := provider.BeginAuth(state)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to begin Google auth"})
	}

	authURL, err := sess.GetAuthURL()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to get Google auth URL"})
	}

	if err := gothic.StoreInSession(provider.Name(), sess.Marshal(), c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to persist OAuth session"})
	}

	if redirectURI != "" {
		c.Response().Header().Set("Set-Cookie", "oauth_redirect_uri="+url.QueryEscape(redirectURI)+"; Path=/; HttpOnly; SameSite=Lax; Max-Age=600")
	}

	return c.Redirect(http.StatusFound, authURL)
}

func (h *UserHandler) GoogleOAuthCallback(c *echo.Context) error {
	gUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Google Authentication failed"})
	}

	ctx := c.Request().Context()
	user, err := h.usecase.GetByEmail(ctx, gUser.Email)
	if err != nil {
		randomPassword := utils.GenerateULID()

		user, err = h.usecase.Register(
			ctx,
			gUser.Email,
			randomPassword,
			gUser.Name,
			"",
			"",
			"",
			"",
			"",
			"",
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to register user"})
		}
	}

	if h.refreshUsecase != nil {
		pair, err := h.refreshUsecase.IssueTokenPair(ctx, user)
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

func (h *UserHandler) GoogleMobileLogin(c *echo.Context) error {
	var req GoogleOAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Invalid request body"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httperrors.Response{Error: "Validation failed"})
	}

	ctx := c.Request().Context()

	info, err := verifyGoogleIDToken(ctx, req.IDToken, h.cfg.GoogleOAuth.ClientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Invalid Google token"})
	}

	user, err := h.usecase.GetByEmail(ctx, info.Email)
	if err != nil {
		randomPassword := utils.GenerateULID()
		user, err = h.usecase.Register(
			ctx,
			info.Email,
			randomPassword,
			info.Username,
			"",
			"",
			"",
			"",
			"",
			"",
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, httperrors.Response{Error: "Failed to register user"})
		}
	}

	if h.refreshUsecase != nil {
		pair, err := h.refreshUsecase.IssueTokenPair(ctx, user)
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
		req.Username,
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
