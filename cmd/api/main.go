// Package main is the entry point for the application.
package main

import (
	"context"
	"database/sql"
	"golang-dining-ordering/config"
	"golang-dining-ordering/internal/routes"
	"golang-dining-ordering/pkg/middleware"
	authHandler "golang-dining-ordering/services/auth/handler"
	authRepo "golang-dining-ordering/services/auth/repository"
	authRoutes "golang-dining-ordering/services/auth/routes"
	authService "golang-dining-ordering/services/auth/service"
	mngHandlers "golang-dining-ordering/services/management/handlers"
	mngRepos "golang-dining-ordering/services/management/repository"
	mngRoutes "golang-dining-ordering/services/management/routes"
	mngServices "golang-dining-ordering/services/management/services"
	"log"
	"log/slog"
	"net/http"
	"os"

	authDB "golang-dining-ordering/services/auth/db/generated"
	managementDB "golang-dining-ordering/services/management/db/generated"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
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

	e := echo.New()
	e.Use(middleware.RequestLogger(logger))

	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	routes.AddSwaggerRoutes(e)

	setupAuth(e, &cfg, logger)
	setupManagement(e, &cfg, logger)

	logger.Info("starting server on address " + cfg.DineHTTPAddress)

	err = e.Start(cfg.DineHTTPAddress)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}

func setupAuth(e *echo.Echo, cfg *config.AppConfig, logger *slog.Logger) {
	authConn, err := sql.Open("postgres", cfg.DineAuthDBURI)
	if err != nil {
		logger.Error("failed to prepare database connection", "error", err)

		return
	}

	err = authConn.PingContext(context.Background())
	if err != nil {
		logger.Error("failed to connect to auth database", "error", err)

		return
	}

	authQueries := authDB.New(authConn)

	usersRepo := authRepo.NewRepository(authQueries)
	authConfig := &authService.Config{
		Secret:                   cfg.DineAuthSecret,
		TokenValidSeconds:        cfg.DineTokenValidSeconds,
		RefreshTokenValidSeconds: cfg.DineRefreshTokenValidSeconds,
	}
	authService := authService.NewAuthService(authConfig, usersRepo)
	authHandler := authHandler.NewAuthHandler(logger, authService)

	authRoutes.AddRoutes(context.Background(), e, authHandler)
}

func setupManagement(e *echo.Echo, cfg *config.AppConfig, _ *slog.Logger) {
	db, _ := sql.Open("postgres", cfg.DineManagementDBURI)

	queries := managementDB.New(db)

	restRepo := mngRepos.NewRestaurantRepository(db, queries)
	restService := mngServices.NewRestaurantService(restRepo)
	restHandler := mngHandlers.NewRestaurantsHandler(restService)

	menuRepo := mngRepos.NewMenuRepository(db, queries)
	menuSvc := mngServices.NewMenuService(menuRepo)
	menuHandler := mngHandlers.NewMenuHandler(menuSvc)

	mngRoutes.AddRrestaurantRoutes(e, restHandler,
		cfg.DineAuthorizeEndpoint,
	)

	mngRoutes.AddMenuRoutes(e, menuHandler, cfg.DineAuthorizeEndpoint)
}
