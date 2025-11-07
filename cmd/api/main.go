// Package main is the entry point for the application.
package main

import (
	"context"
	"database/sql"
	"golang-dining-ordering/config"
	"golang-dining-ordering/internal/routes"
	authHandler "golang-dining-ordering/services/auth/handler"
	authRepo "golang-dining-ordering/services/auth/repository"
	authRoutes "golang-dining-ordering/services/auth/routes"
	authService "golang-dining-ordering/services/auth/service"
	"log"
	"log/slog"
	"net/http"
	"os"

	authDB "golang-dining-ordering/services/auth/db/generated"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	// DefaultTokenValidHours is the default duration (in hours) that an access token is valid.
	DefaultTokenValidHours = 168
	// DefaultRefreshTokenValidHours is the default duration (in hours) that a refresh token is valid.
	DefaultRefreshTokenValidHours = 336
)

func main() {
	var cfg config.AppConfig

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Panic("failed to load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))

	conn, err := sql.Open("postgres", cfg.DineDBURI)
	if err != nil {
		logger.Error("failed to prepare database connection", "error", err)

		return
	}

	err = conn.PingContext(context.Background())
	if err != nil {
		logger.Error("failed to connect to database", "error", err)

		return
	}

	queries := authDB.New(conn)

	e := echo.New()

	usersRepo := authRepo.NewUserRepository(queries)
	authConfig := &authService.Config{
		Secret:                   cfg.DineAuthSecret,
		TokenValidSeconds:        cfg.DineTokenValidSeconds,
		RefreshTokenValidSeconds: cfg.DineRefreshTokenValidSeconds,
	}
	authService := authService.NewAuthService(authConfig, usersRepo)

	authHandler := authHandler.NewAuthHandler(logger, authService)

	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	routes.AddSwaggerRoutes(e)

	authRoutes.AddRoutes(context.Background(), e, authHandler)

	logger.Info("starting server on address " + cfg.DineHTTPAddress)

	err = e.Start(cfg.DineHTTPAddress)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}
