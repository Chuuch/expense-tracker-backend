package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	applogger "github.com/chuuch/expense-tracker-backend/pkg/logger"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

func TestLoggerMiddleware_PassesThroughError(t *testing.T) {
	applogger.Log = zap.NewNop()
	
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedErr := errors.New("next failed")
	next := func(c *echo.Context) error {
		return expectedErr
	}

	err := authhttp.LoggerMiddleware()(next)(c)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected next error, got %v", err)
	}
}
