package app

import (
	httpserver "boton-back/internal/app/http-server"
	"boton-back/internal/config"
	"boton-back/internal/handlers"
	"boton-back/internal/lib/jwt"
	"boton-back/internal/middlewares"
	"boton-back/internal/repository/postgres"
	"boton-back/internal/repository/redis"
	"boton-back/internal/routes"
	"boton-back/internal/services"
	"context"
	"log/slog"
)

type App struct {
	HTTPServer *httpserver.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	storage, err := postgres.NewPostgres(ctx, cfg.Database.PostgresConn)
	if err != nil {
		panic(err)
	}

	redisDB, err := redis.InitRedis(cfg.Redis.RedisConn, cfg.Redis.RedisUsername, cfg.Redis.RedisPassword, cfg.Redis.RedisPassword, cfg.Redis.MaxRetries, cfg.Redis.Timeout, cfg.JWT.RefreshExpirationDays)
	if err != nil {
		panic(err)
	}

	jwtGenerator := jwt.NewGenerator(cfg.JWT.Secret, cfg.JWT.AccessExpirationMinutes, cfg.JWT.RefreshExpirationDays)

	authService := services.NewAuthService(log, jwtGenerator, storage, redisDB)
	userService := services.NewUserService(log, storage)

	authHandler := handlers.NewAuthHandler(log, authService)
	userHandler := handlers.NewUserHandler(log, userService)

	authMiddleware := middlewares.NewAuthMiddleware(jwtGenerator)

	r := routes.InitRoutes(authHandler, userHandler, authMiddleware)

	server := httpserver.NewServer(log, cfg.Server.AuthAddress, cfg.Server.AuthTimeout, r)

	return &App{
		HTTPServer: server,
	}
}
