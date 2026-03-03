package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	goalhttp "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/goals/domain"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/goals/usecase/interfaces/mocks"
	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
	"go.uber.org/mock/gomock"
)

func newGoalEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewEchoValidator()
	return e
}

func withAuthUser(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), authhttp.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestCreateGoal_Success_Returns201(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalUC := mocks.NewMockGoalUsecase(ctrl)
	h := goalhttp.NewGoalHandler(goalUC)
	e := newGoalEcho()

	body := map[string]any{
		"name":          "Emergency Fund",
		"currency":      "usd",
		"target_amount": 100000,
		"target_date":   "2026-12-31",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/goals", bytes.NewReader(b))
	req = withAuthUser(req, "u1")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	targetDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	goalUC.EXPECT().
		CreateGoal(
			gomock.Any(),
			"u1",
			"Emergency Fund",
			"usd",
			int64(100000),
			&targetDate,
		).
		Return(&domain.Goal{
			ID:           "g1",
			UserID:       "u1",
			Name:         "Emergency Fund",
			Currency:     "USD",
			TargetAmount: 100000,
			TargetDate:   &targetDate,
			Status:       domain.GoalStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil).
		Times(1)

	if err := h.CreateGoal(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateGoal_Unauthorized_Returns401(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalUC := mocks.NewMockGoalUsecase(ctrl)
	h := goalhttp.NewGoalHandler(goalUC)
	e := newGoalEcho()

	body := `{"name":"Emergency Fund","currency":"usd","target_amount":100000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/goals", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	goalUC.EXPECT().
		CreateGoal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	if err := h.CreateGoal(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetGoalByID_NotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalUC := mocks.NewMockGoalUsecase(ctrl)
	h := goalhttp.NewGoalHandler(goalUC)
	e := newGoalEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/goals/g_missing", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/goals/:id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "g_missing"}})

	goalUC.EXPECT().
		GetGoalByID(gomock.Any(), "u1", "g_missing").
		Return(nil, usecase.ErrGoalNotFound).
		Times(1)

	if err := h.GetGoalByID(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListGoals_Success_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalUC := mocks.NewMockGoalUsecase(ctrl)
	h := goalhttp.NewGoalHandler(goalUC)
	e := newGoalEcho()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/goals?status=active&page_size=10&offset=0",
		nil,
	)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	status := domain.GoalStatusActive

	goalUC.EXPECT().
		ListGoals(gomock.Any(), gomock.AssignableToTypeOf(interfaces.GoalListFilter{})).
		DoAndReturn(func(_ context.Context, f interfaces.GoalListFilter) ([]*domain.Goal, error) {
			if f.UserID != "u1" {
				t.Fatalf("expected userID u1, got %s", f.UserID)
			}
			if f.Status == nil || *f.Status != status {
				t.Fatalf("expected status active, got %+v", f.Status)
			}
			if f.PageSize != 10 || f.Offset != 0 {
				t.Fatalf("expected page_size/offset 10/0, got %d/%d", f.PageSize, f.Offset)
			}
			return []*domain.Goal{
				{
					ID:           "g1",
					UserID:       "u1",
					Name:         "Emergency Fund",
					Currency:     "USD",
					TargetAmount: 100000,
					Status:       domain.GoalStatusActive,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			}, nil
		}).
		Times(1)

	if err := h.ListGoals(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAddContribution_Success_Returns201(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalUC := mocks.NewMockGoalUsecase(ctrl)
	h := goalhttp.NewGoalHandler(goalUC)
	e := newGoalEcho()

	body := `{"amount":2500,"contribution_date":"2026-03-01","note":"first deposit"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/goals/g1/contributions", bytes.NewBufferString(body))
	req = withAuthUser(req, "u1")
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/goals/:id/contributions")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "g1"}})

	contribDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	goalUC.EXPECT().
		AddContribution(
			gomock.Any(),
			"u1",
			"g1",
			int64(2500),
			contribDate,
			"first deposit",
		).
		Return(&domain.GoalContribution{
			ID:               "c1",
			GoalID:           "g1",
			UserID:           "u1",
			Amount:           2500,
			ContributionDate: contribDate,
			Note:             "first deposit",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}, nil).
		Times(1)

	if err := h.AddContribution(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetGoalProgress_Success_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goalUC := mocks.NewMockGoalUsecase(ctrl)
	h := goalhttp.NewGoalHandler(goalUC)
	e := newGoalEcho()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/goals/g1/progress", nil)
	req = withAuthUser(req, "u1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/goals/:id/progress")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "g1"}})

	now := time.Now()

	goalUC.EXPECT().
		GetGoalProgress(gomock.Any(), "u1", "g1").
		Return(&interfaces.GoalProgress{
			GoalID:             "g1",
			TargetAmount:       100000,
			SavedAmount:        25000,
			RemainingAmount:    75000,
			ProgressPercentage: 25,
			IsCompleted:        false,
			LastUpdatedAt:      now,
		}, nil).
		Times(1)

	if err := h.GetGoalProgress(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}
