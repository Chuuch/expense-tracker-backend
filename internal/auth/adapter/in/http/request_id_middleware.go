package http

import (
	"context"
	"strings"

	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
)

const requestIDHeader = "X-Request-ID"

func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestID := strings.TrimSpace(c.Request().Header.Get(requestIDHeader))
			if requestID == "" {
				requestID = utils.GenerateULID()
			}

			c.Response().Header().Set(requestIDHeader, requestID)
			c.Request().Header.Set(requestIDHeader, requestID)
			c.Set(string(RequestIDKey), requestID)

			ctx := context.WithValue(c.Request().Context(), RequestIDKey, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
