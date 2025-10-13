package routes

import (
	"context"
	"golang-dining-ordering/internal/handlers"

	"github.com/labstack/echo/v4"
)

func AddAuthRoutes(ctx context.Context, e *echo.Echo, h *handlers.AuthHandler) {
	auth := e.Group("/api/v1/auth")

	auth.POST("/signup", h.HandleSignUp)
	auth.POST("/signin", h.HandleSignIn)
	auth.POST("/refresh", h.HandleRefreshToken)
}
