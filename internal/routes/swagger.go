// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	swagui "github.com/swaggest/swgui/v3"
)

// AddSwaggerRoutes registers Swagger UI and specification endpoints.
func AddSwaggerRoutes(e *echo.Echo) {
	e.File("/swagger.yml", "api/openapi-spec/swagger.yml")

	swaggerHandler := swagui.NewHandler("Dining Ordering API", "/swagger.yml", "/docs")
	e.GET("/docs/*", echo.WrapHandler(swaggerHandler))

	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/docs/")
	})
}
