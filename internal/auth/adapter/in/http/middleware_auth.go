package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/labstack/echo/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

const (
	AuthHeader = "Authorization"
	AuthScheme = "Bearer"
)

func AuthMiddleware(tokenUsecase interfaces.TokenUsecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get(AuthHeader)
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing authorization header"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], AuthScheme) {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid authorization header"})
			}

			token := parts[1]
			userID, err := tokenUsecase.VerifyToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
			}

			c.Set(string(UserIDKey), userID)

			newCtx := context.WithValue(c.Request().Context(), UserIDKey, userID)
			c.SetRequest(c.Request().WithContext(newCtx))

			return next(c)
		}
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
