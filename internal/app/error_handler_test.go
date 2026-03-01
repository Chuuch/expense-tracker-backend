package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestHTTPErrorHandler_InternalError_Normalized500(t *testing.T) {
	a := newRouterSmokeApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-ID", "req-1")
	c := a.echo.NewContext(req, rec)

	a.httpErrorHandler(c, errors.New("db exploded"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	if body["error"] != "Internal server error" {
		t.Fatalf("expected internal error message, got %v", body["error"])
	}
	if body["request_id"] != "req-1" {
		t.Fatalf("expected request_id req-1, got %v", body["request_id"])
	}
}

func TestHTTPErrorHandler_HTTPErrorPreserves4xxMessage(t *testing.T) {
	a := newRouterSmokeApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	c := a.echo.NewContext(req, rec)

	a.httpErrorHandler(c, echo.NewHTTPError(http.StatusBadRequest, "Validation failed"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	if body["error"] != "Validation failed" {
		t.Fatalf("expected validation message, got %v", body["error"])
	}
}

func TestHTTPErrorHandler_HTTPErrorHides5xxMessage(t *testing.T) {
	a := newRouterSmokeApp()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	c := a.echo.NewContext(req, rec)

	a.httpErrorHandler(c, echo.NewHTTPError(http.StatusInternalServerError, "sensitive details"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	if body["error"] != "Internal server error" {
		t.Fatalf("expected sanitized message, got %v", body["error"])
	}
}
