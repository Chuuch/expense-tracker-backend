package app

import (
	analyticshttp "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/in/http"
	analyticspostgres "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/out/postgres"
	analyticsusecase "github.com/chuuch/expense-tracker-backend/internal/analytics/usecase"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

func buildAnalyticsModule(q postgresdb.Querier) *analyticshttp.AnalyticsHandler {
	analyticsRepo := analyticspostgres.NewAnalyticsRepository(q)
	analyticsUC := analyticsusecase.NewAnalyticsUsecase(analyticsRepo)
	return analyticshttp.NewAnalyticsHandler(analyticsUC)
}
