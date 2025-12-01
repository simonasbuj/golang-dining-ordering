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
	mngStorage "golang-dining-ordering/services/management/storage/local"
	ordersHandlers "golang-dining-ordering/services/orders/handlers"
	"golang-dining-ordering/services/orders/paymentproviders"
	ordersRepo "golang-dining-ordering/services/orders/repository"
	ordersRoutes "golang-dining-ordering/services/orders/routes"
	ordersServices "golang-dining-ordering/services/orders/services"
	"log"
	"log/slog"
	"net/http"
	"os"

	authDB "golang-dining-ordering/services/auth/db/generated"
	managementDB "golang-dining-ordering/services/management/db/generated"
	ordersDB "golang-dining-ordering/services/orders/db/generated"

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
	routes.AddFrontendRoutes(e)

	setupAuth(e, &cfg, logger)
	setupManagement(e, &cfg, logger)
	setupOrders(e, &cfg, logger)

	logger.Info("starting server on address " + cfg.HTTPAddress)

	err = e.Start(cfg.HTTPAddress)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}

func setupAuth(e *echo.Echo, cfg *config.AppConfig, logger *slog.Logger) {
	authConn, err := sql.Open("postgres", cfg.AuthDBURI)
	if err != nil {
		logger.Error("failed to prepare database connection", "error", err)
		os.Exit(1)
	}

	err = authConn.PingContext(context.Background())
	if err != nil {
		logger.Error("failed to connect to auth database", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to auth db")

	authQueries := authDB.New(authConn)

	usersRepo := authRepo.NewRepository(authQueries)
	authConfig := &authService.Config{
		Secret:                   cfg.AuthSecret,
		TokenValidSeconds:        cfg.TokenValidSeconds,
		RefreshTokenValidSeconds: cfg.RefreshTokenValidSeconds,
	}
	authService := authService.NewAuthService(authConfig, usersRepo)
	authHandler := authHandler.NewAuthHandler(logger, authService)

	authRoutes.AddRoutes(context.Background(), e, authHandler)
}

func setupManagement(e *echo.Echo, cfg *config.AppConfig, logger *slog.Logger) {
	db, err := sql.Open("postgres", cfg.ManagementDBURI)
	if err != nil {
		logger.Error("failed to prepare database connection", "error", err)
		os.Exit(1)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		logger.Error("failed to connect to management database", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to management db")

	queries := managementDB.New(db)

	restRepo := mngRepos.NewRestaurantRepository(db, queries)
	restService := mngServices.NewRestaurantService(restRepo)
	restHandler := mngHandlers.NewRestaurantsHandler(restService)

	menuRepo := mngRepos.NewMenuRepository(db, queries)
	storage := mngStorage.NewLocalStorage(cfg.MaxImageSizeBytes, cfg.UploadsDirectory)
	menuSvc := mngServices.NewMenuService(menuRepo, restRepo, storage)
	menuHandler := mngHandlers.NewMenuHandler(menuSvc)

	mngRoutes.AddRestaurantRoutes(e, restHandler,
		cfg.AuthorizeEndpoint,
	)

	mngRoutes.AddMenuRoutes(e, menuHandler, cfg.AuthorizeEndpoint)
}

func setupOrders(e *echo.Echo, cfg *config.AppConfig, logger *slog.Logger) {
	db, err := sql.Open("postgres", cfg.ManagementDBURI)
	if err != nil {
		logger.Error("failed to prepare orders db connection", "error", err)
		os.Exit(1)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		logger.Error("failed to connect to orders database", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to orders db")

	queries := ordersDB.New(db)

	ordersRepo := ordersRepo.New(queries)
	ordersSvc := ordersServices.NewOrdersService(ordersRepo)
	ordersHandler := ordersHandlers.NewOrdersHandler(ordersSvc)

	paymentsProvider := paymentproviders.NewStripePaymentProvider(
		cfg.StripeSecretKey,
		cfg.StripeWebhookSecret,
	)
	paymentsSvc := ordersServices.NewPaymentsService(ordersRepo, paymentsProvider)
	paymentsHandler := ordersHandlers.NewPaymentsHandler(paymentsSvc)
	ordersRoutes.AddOrdersRoutes(e, ordersHandler, paymentsHandler, cfg.AuthorizeEndpoint)
}
