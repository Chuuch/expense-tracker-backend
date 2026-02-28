package http

import (
	"time"

	"github.com/chuuch/expense-tracker-backend/pkg/logger"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

func LoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			if err != nil {
				c.Logger().Error(err.Error())
			}

			stop := time.Now()
			req := c.Request()

			userID, _ := c.Get(string(UserIDKey)).(string)
			if userID == "" {
				userID = "unknown"
			}

			logger.Log.Info("HTTP Request",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", c.Request().Response.StatusCode),
				zap.Duration("latency", stop.Sub(start)),
				zap.String("ip", c.RealIP()),
				zap.String("user_id", userID),
			)
			return err
		}
	}
}
