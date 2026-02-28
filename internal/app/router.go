package app

import (
	"net/http"

	authHttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/labstack/echo/v5"
)

func (a *App) registerRoutes() {
	a.echo.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	auth := a.echo.Group("/api/v1/auth")
	auth.POST("/register", a.userHandler.Register)
	auth.POST("/login", a.userHandler.Login)

	users := a.echo.Group("/api/v1/users", authHttp.AuthMiddleware(a.tokenUsecase))
	users.GET("/:id", a.userHandler.GetByID)
	users.PUT("/:id", a.userHandler.UpdateUser)
	users.DELETE("/:id", a.userHandler.DeleteUser)

}
