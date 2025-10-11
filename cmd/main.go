package main

import (
	"context"
	"database/sql"
	"fmt"
	db "golang-dining-ordering/internal/db/generated"
	"golang-dining-ordering/internal/handlers"
	"golang-dining-ordering/internal/repository"
	"golang-dining-ordering/internal/routes"
	"golang-dining-ordering/internal/services"
	"golang-dining-ordering/pkg/utils/env"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	// env vars
	dbUri := env.GetString("DB_URI", "postgres://postgres:postgres@localhost:5432/dining?sslmode=disable")
	httpPort := env.GetString("HTTP_PORT", ":42069")

	// logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))

	// database
	conn, err := sql.Open("postgres", dbUri)
	if err != nil {
		logger.Error("failed to prepare database connection", "error", err)
		return
	}

	if err := conn.Ping(); err != nil {
		logger.Error("failed to connect to database", "error", err)
		return
	}

	queries := db.New(conn)

	// dependency injection
	e := echo.New()

	usersRepo := repository.NewUserRepository(queries)
	usersService := services.NewUserService(usersRepo)

	authHandler := handlers.NewAuthHandler(logger, usersService)

	// register reoutes
	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	routes.AddSwaggerRoutes(e)

	routes.AddAuthRoutes(context.Background(), e, authHandler)

	// start server
	logger.Info(fmt.Sprintf("starting server on port %s", httpPort))
	err = e.Start(httpPort)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}
