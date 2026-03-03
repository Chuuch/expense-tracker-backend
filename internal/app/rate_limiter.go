package app

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

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
