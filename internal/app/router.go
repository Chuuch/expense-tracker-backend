package app

import (
	"context"
	"net/http"
	"time"

	authHttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func (a *App) registerRoutes() {
	a.echo.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	a.echo.GET("/health/live", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	a.echo.GET("/health/ready", a.readiness)

	auth := a.echo.Group("/api/v1/auth")
	auth.Use(a.authRateLimiter())
	auth.POST("/register", a.userHandler.Register)
	auth.POST("/login", a.userHandler.Login)
	auth.POST("/refresh", a.userHandler.Refresh)
	auth.POST("/logout", a.userHandler.Logout, authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))

	users := a.echo.Group("/api/v1/users", authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	users.GET("/:id", a.userHandler.GetByID, authHttp.RequireSelfOrRoles(domain.RoleAdmin, domain.RoleSupport))
	users.PUT("/:id", a.userHandler.UpdateUser, authHttp.RequireSelfOrRoles(domain.RoleAdmin, domain.RoleSupport))
	users.DELETE("/:id", a.userHandler.DeleteUser, authHttp.RequireSelfOrRoles(domain.RoleAdmin))

	expenses := a.echo.Group("/api/v1/expenses", authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	expenses.POST("", a.expenseHandler.CreateExpense)
	expenses.GET("", a.expenseHandler.ListExpenses)
	expenses.GET("/:id", a.expenseHandler.GetExpenseByID)
	expenses.PUT("/:id", a.expenseHandler.UpdateExpense)
	expenses.DELETE("/:id", a.expenseHandler.DeleteExpense)
}

func (a *App) readiness(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if a.db == nil || a.redis == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
	}

	if err := a.db.PingContext(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
	}
	if err := a.redis.Ping(ctx).Err(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}

func (a *App) authRateLimiter() echo.MiddlewareFunc {
	ratePerSecond := float64(a.cfg.Auth.RateLimitRequests) / a.cfg.Auth.RateLimitWindow.Seconds()
	if ratePerSecond <= 0 {
		ratePerSecond = 1
	}

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      ratePerSecond,
			Burst:     a.cfg.Auth.RateLimitRequests,
			ExpiresIn: a.cfg.Auth.RateLimitWindow,
		}),
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(_ *echo.Context, err error) error {
			return err
		},
		DenyHandler: func(c *echo.Context, _ string, _ error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too many requests"})
		},
	})
}
