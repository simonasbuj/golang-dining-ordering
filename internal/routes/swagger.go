// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"github.com/labstack/echo/v4"
	swagui "github.com/swaggest/swgui/v3"
)

// AddSwaggerRoutes registers Swagger UI and specification endpoints.
func AddSwaggerRoutes(e *echo.Echo) {
	e.Static("/openapi-spec", "api/openapi-spec")

	swaggerHandler := swagui.NewHandler(
		"Dining Ordering API",
		"/openapi-spec/openapi-spec.yml",
		"/docs",
	)
	e.GET("/docs*", echo.WrapHandler(swaggerHandler))
}
