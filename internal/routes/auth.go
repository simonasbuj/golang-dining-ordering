// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"context"

	"github.com/labstack/echo/v4"
	"golang-dining-ordering/internal/handlers"
)

// AddAuthRoutes registers authentication-related HTTP routes.
func AddAuthRoutes(_ context.Context, e *echo.Echo, h *handlers.AuthHandler) {
	auth := e.Group("/api/v1/auth")

	auth.POST("/signup", h.HandleSignUp)
	auth.POST("/signin", h.HandleSignIn)
	auth.POST("/refresh", h.HandleRefreshToken)
}
