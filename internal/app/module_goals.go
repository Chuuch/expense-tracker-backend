package app

import (
	goalhttp "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/in/http"
	goalPostgres "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/out/postgres"
	goalusecase "github.com/chuuch/expense-tracker-backend/internal/goals/usecase"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
)

func buildGoalsModule(q postgresdb.Querier) *goalhttp.GoalHandler {
	goalRepo := goalPostgres.NewGoalRepository(q)
	contributionRepo := goalPostgres.NewGoalContributionRepository(q)

	goalUsecase := goalusecase.NewGoalUsecase(goalRepo, contributionRepo)
	return goalhttp.NewGoalHandler(goalUsecase)
}
