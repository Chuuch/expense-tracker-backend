package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/labstack/echo/v5"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	RoleKey      contextKey = "role"
	RequestIDKey contextKey = "request_id"
)

const (
	AuthHeader = "Authorization"
	AuthScheme = "Bearer"
)

func AuthMiddleware(tokenUsecase interfaces.TokenUsecase, blacklist interfaces.AccessTokenBlacklist) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token, ok := extractBearerToken(c.Request().Header.Get(AuthHeader))
			if !ok {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing authorization header"})
			}

			if blacklist != nil {
				blacklisted, err := blacklist.Contains(c.Request().Context(), token)
				if err != nil || blacklisted {
					return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
				}
			}

			claims, err := tokenUsecase.VerifyToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
			}

			c.Set(string(UserIDKey), claims.UserID)
			c.Set(string(RoleKey), string(claims.Role)) // store as string consistently

			ctx := context.WithValue(c.Request().Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, string(claims.Role))
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func RequireSelfOrRoles(allowed ...domain.UserRole) echo.MiddlewareFunc {
	allowedSet := make(map[domain.UserRole]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			pathID := c.Param("id")
			userID, _ := c.Get(string(UserIDKey)).(string)
			roleStr, _ := c.Get(string(RoleKey)).(string)
			role := domain.UserRole(roleStr)

			if pathID == userID {
				return next(c)
			}
			if _, ok := allowedSet[role]; ok {
				return next(c)
			}

			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		}
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetRoleFromContext(ctx context.Context) (domain.UserRole, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return domain.UserRole(role), ok
}

func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

func extractBearerToken(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], AuthScheme) {
		return "", false
	}

	return strings.TrimSpace(parts[1]), parts[1] != ""
}
