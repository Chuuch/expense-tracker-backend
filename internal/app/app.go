package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	analyticshttp "github.com/chuuch/expense-tracker-backend/internal/analytics/adapter/in/http"
	authHttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	expenseHttp "github.com/chuuch/expense-tracker-backend/internal/expenses/adapter/in/http"
	goalhttp "github.com/chuuch/expense-tracker-backend/internal/goals/adapter/in/http"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	"github.com/chuuch/expense-tracker-backend/utils"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type App struct {
	echo                 *echo.Echo
	cfg                  *config.Config
	log                  *zap.Logger
	userHandler          *authHttp.UserHandler
	expenseHandler       *expenseHttp.ExpenseHandler
	goalHandler          *goalhttp.GoalHandler
	analyticsHandler     *analyticshttp.AnalyticsHandler
	tokenUsecase         interfaces.TokenUsecase
	accessTokenBlacklist interfaces.AccessTokenBlacklist
	db                   *sql.DB
	redis                *redis.Client
}

func NewApp(
	cfg *config.Config,
	log *zap.Logger,
	userHandler *authHttp.UserHandler,
	expenseHandler *expenseHttp.ExpenseHandler,
	goalHandler *goalhttp.GoalHandler,
	analyticsHandler *analyticshttp.AnalyticsHandler,
	tokenUsecase interfaces.TokenUsecase,
	accessTokenBlacklist interfaces.AccessTokenBlacklist,
	db *sql.DB,
	redisClient *redis.Client,
) *App {
	e := echo.New()
	e.Validator = utils.NewEchoValidator()

	application := &App{
		echo:                 e,
		cfg:                  cfg,
		log:                  log,
		userHandler:          userHandler,
		expenseHandler:       expenseHandler,
		goalHandler:          goalHandler,
		analyticsHandler:     analyticsHandler,
		tokenUsecase:         tokenUsecase,
		accessTokenBlacklist: accessTokenBlacklist,
		db:                   db,
		redis:                redisClient,
	}

	e.HTTPErrorHandler = application.httpErrorHandler

	return application
}

func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.echo.Use(middleware.Recover())
	a.echo.Use(authHttp.RequestIDMiddleware())
	a.echo.Use(authHttp.LoggerMiddleware())

	if a.cfg.App.AppEnv == "production" {
		a.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection: "1; mode=block",
			ContentTypeNosniff: "nosniff",
			XFrameOptions: "DENY",
			HSTSMaxAge: 31536000,
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self' wss:; frame-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; sandbox allow-same-origin allow-scripts allow-forms allow-top-navigation allow-popups allow-modals allow-pointer-lock allow-orientation-lock allow-popunder; media-src 'self'; worker-src 'self' blob:; child-src 'self' blob:; prefetch-src 'self' blob:; manifest-src 'self';",
		}))
	}

	if len(a.cfg.Server.AllowedOrigins) > 0 {
		a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: a.cfg.Server.AllowedOrigins,
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
			},
			AllowHeaders: []string{
				echo.HeaderOrigin,
				echo.HeaderContentType,
				echo.HeaderAccept,
				echo.HeaderAuthorization,
				"X-Request-ID",
			},
			ExposeHeaders: []string{"X-Request-ID"},
		}))
	}
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
