// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"context"
	"golang-dining-ordering/services/auth/handler"

	"github.com/labstack/echo/v4"
)

// AddRoutes registers authentication-related HTTP routes.
func AddRoutes(_ context.Context, e *echo.Echo, h *handler.Handler) {
	auth := e.Group("/api/v1/auth")

	auth.POST("/signup", h.HandleSignUp)
	auth.POST("/signin", h.HandleSignIn)
	auth.POST("/refresh", h.HandleRefreshToken)
}
