package app

import (
	"fmt"

	authhttp "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/in/http"
	authasynq "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/asynq"
	authpostgres "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/postgres"
	authredis "github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/redis"
	"github.com/chuuch/expense-tracker-backend/internal/auth/adapter/out/token/paseto"
	authusecase "github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/internal/platform/config"
	postgresdb "github.com/chuuch/expense-tracker-backend/internal/storage/postgres/sqlc"
	"github.com/gorilla/sessions"
	"github.com/hibiken/asynq"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/redis/go-redis/v9"
)

type authModule struct {
	handler              *authhttp.UserHandler
	tokenUsecase         interfaces.TokenUsecase
	accessTokenBlacklist interfaces.AccessTokenBlacklist
	asynqClient          *asynq.Client
}

func buildAuthModule(cfg *config.Config, q postgresdb.Querier, redisClient *redis.Client) (*authModule, error) {
	if cfg.GoogleOAuth.ClientID != "" && cfg.GoogleOAuth.ClientSecret != "" && cfg.GoogleOAuth.CallbackURL != "" {
		goth.UseProviders(
			google.New(
				cfg.GoogleOAuth.ClientID,
				cfg.GoogleOAuth.ClientSecret,
				cfg.GoogleOAuth.CallbackURL,
				"email",
				"profile",
			),
		)
		gothic.Store = sessions.NewCookieStore([]byte(cfg.Auth.PasetoSymmetricKey))
	}

	userRepo := authpostgres.NewUserRepository(q)

	asynqRedisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	asynqClient := asynq.NewClient(asynqRedisOpt)
	verificationEnqueuer := authasynq.NewVerificationEmailEnqueuer(asynqClient)

	tokenUsecase, err := paseto.NewPasetoUsecase(cfg.Auth.PasetoSymmetricKey)
	if err != nil {
		return nil, err
	}

	refreshRepo := authpostgres.NewRefreshTokenRepository(q)
	userUsecase := authusecase.NewUserUsecase(userRepo, verificationEnqueuer)
	refreshUsecase := authusecase.NewRefreshTokenUsecase(
		refreshRepo,
		userRepo,
		tokenUsecase,
		cfg,
	)

	blacklist := authredis.NewAccessTokenBlacklistRepository(redisClient)
	handler := authhttp.NewUserHandlerWithRefresh(userUsecase, tokenUsecase, refreshUsecase, cfg, blacklist)

	return &authModule{
		handler:              handler,
		tokenUsecase:         tokenUsecase,
		accessTokenBlacklist: blacklist,
		asynqClient:          asynqClient,
	}, nil
}
