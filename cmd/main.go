// Package main is the entry point for the application.
package main

import (
	"context"
	"database/sql"
	"golang-dining-ordering/internal/handlers"
	"golang-dining-ordering/internal/repository"
	"golang-dining-ordering/internal/routes"
	"golang-dining-ordering/internal/services"
	"golang-dining-ordering/pkg/utils/env"
	"log/slog"
	"net/http"
	"os"

	db "golang-dining-ordering/internal/db/generated"

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
	// env vars
	dbURI := env.GetString(
		"DB_URI",
		"postgres://postgres:postgres@localhost:5432/dining?sslmode=disable",
	)
	httpPort := env.GetString("HTTP_PORT", ":42069")
	authSecret := env.GetString("AUTH_SECRET", "my-auth-secret")
	tokenValidHours := env.GetInt("TOKEN_VALID_HOURS", DefaultTokenValidHours)
	refreshTokenValidHours := env.GetInt("REFRESH_TOKEN_VALID_HOURS", DefaultRefreshTokenValidHours)

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))

	// database
	conn, err := sql.Open("postgres", dbURI)
	if err != nil {
		logger.Error("failed to prepare database connection", "error", err)

		return
	}

	err = conn.PingContext(context.Background())
	if err != nil {
		logger.Error("failed to connect to database", "error", err)

		return
	}

	queries := db.New(conn)

	// dependency injection
	e := echo.New()

	usersRepo := repository.NewUserRepository(queries)
	authConfig := &services.AuthConfig{
		Secret:                 authSecret,
		TokenValidHours:        tokenValidHours,
		RefreshTokenValidHours: refreshTokenValidHours,
	}
	authService := services.NewAuthService(authConfig, usersRepo)

	authHandler := handlers.NewAuthHandler(logger, authService)

	// register reoutes
	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	routes.AddSwaggerRoutes(e)

	routes.AddAuthRoutes(context.Background(), e, authHandler)

	// start server
	logger.Info("starting server on port " + httpPort)

	err = e.Start(httpPort)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}
