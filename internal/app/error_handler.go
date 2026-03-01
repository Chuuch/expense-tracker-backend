package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func (a *App) httpErrorHandler(c *echo.Context, err error) {
	if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && resp.Committed {
		return
	}

	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		statusCode = httpErr.Code
		if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
			if msg := strings.TrimSpace(httpErr.Message); msg != "" {
				message = msg
			}
		} else {
			// Never leak internals for server-side errors.
			message = "Internal server error"
		}
	}

	requestID := strings.TrimSpace(c.Response().Header().Get("X-Request-ID"))
	if requestID == "" {
		requestID = strings.TrimSpace(c.Request().Header.Get("X-Request-ID"))
	}

	if a.log != nil {
		a.log.Error("request failed",
			zap.Error(err),
			zap.Int("status", statusCode),
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
			zap.String("request_id", requestID),
		)
	}

	_ = c.JSON(statusCode, errorResponse{
		Error:     message,
		RequestID: requestID,
	})
}
