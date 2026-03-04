package app

import (
	"net/http"

	authHttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/labstack/echo/v5"
)

func (a *App) registerRoutes() {
	a.echo.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	a.echo.GET("/health/live", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	a.echo.GET("/health/ready", a.readiness)

	// AUTH ROUTES
	auth := a.echo.Group("/api/v1/auth")
	auth.Use(a.authRateLimiter())
	auth.POST("/register", a.userHandler.Register)
	auth.POST("/login", a.userHandler.Login)
	auth.POST("/refresh", a.userHandler.Refresh)
	auth.POST("/logout", a.userHandler.Logout, authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	auth.GET("/google", a.userHandler.GoogleOAuthStart)
	auth.GET("/google/callback", a.userHandler.GoogleOAuthCallback)

	// USER ROUTES
	users := a.echo.Group("/api/v1/users", authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	users.GET("/:id", a.userHandler.GetByID, authHttp.RequireSelfOrRoles(domain.RoleAdmin, domain.RoleSupport))
	users.PUT("/:id", a.userHandler.UpdateUser, authHttp.RequireSelfOrRoles(domain.RoleAdmin, domain.RoleSupport))
	users.DELETE("/:id", a.userHandler.DeleteUser, authHttp.RequireSelfOrRoles(domain.RoleAdmin))

	// EXPENSES ROUTES
	expenses := a.echo.Group("/api/v1/expenses", authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	expenses.POST("", a.expenseHandler.CreateExpense)
	expenses.GET("", a.expenseHandler.ListExpenses)
	expenses.GET("/:id", a.expenseHandler.GetExpenseByID)
	expenses.PUT("/:id", a.expenseHandler.UpdateExpense)
	expenses.DELETE("/:id", a.expenseHandler.DeleteExpense)

	// GOALS ROUTES
	goals := a.echo.Group("/api/v1/goals", authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	goals.POST("", a.goalHandler.CreateGoal)
	goals.GET("", a.goalHandler.ListGoals)
	goals.GET("/:id", a.goalHandler.GetGoalByID)
	goals.PATCH("/:id", a.goalHandler.UpdateGoal)
	goals.DELETE("/:id", a.goalHandler.DeleteGoal)
	goals.POST("/:id/contributions", a.goalHandler.AddContribution)
	goals.GET("/:id/contributions", a.goalHandler.ListContributions)
	goals.GET("/:id/progress", a.goalHandler.GetGoalProgress)

	// ANALYTICS ROUTES
	analytics := a.echo.Group("/api/v1/analytics", authHttp.AuthMiddleware(a.tokenUsecase, a.accessTokenBlacklist))
	analytics.GET("/monthly", a.analyticsHandler.GetMonthlySpending)
	analytics.GET("/by-category", a.analyticsHandler.GetSpendingByCategory)
}
