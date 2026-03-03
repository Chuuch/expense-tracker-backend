package app

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

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
