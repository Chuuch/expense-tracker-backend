package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	authHttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	expenseHttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type App struct {
	echo                 *echo.Echo
	cfg                  *config.Config
	log                  *zap.Logger
	userHandler          *authHttp.UserHandler
	expenseHandler       *expenseHttp.ExpenseHandler
	tokenUsecase         interfaces.TokenUsecase
	accessTokenBlacklist interfaces.AccessTokenBlacklist
}

func NewApp(
	cfg *config.Config,
	log *zap.Logger,
	userHandler *authHttp.UserHandler,
	expenseHandler *expenseHttp.ExpenseHandler,
	tokenUsecase interfaces.TokenUsecase,
	accessTokenBlacklist interfaces.AccessTokenBlacklist,
) *App {
	e := echo.New()
	e.Validator = utils.NewEchoValidator()

	return &App{
		echo:                 e,
		cfg:                  cfg,
		log:                  log,
		userHandler:          userHandler,
		expenseHandler:       expenseHandler,
		tokenUsecase:         tokenUsecase,
		accessTokenBlacklist: accessTokenBlacklist,
	}
}

func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.echo.Use(authHttp.LoggerMiddleware())
	a.registerRoutes()

	srv := &http.Server{
		Addr:         a.cfg.Server.Port,
		Handler:      a.echo,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
		IdleTimeout:  a.cfg.Server.IdleTimeout,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
