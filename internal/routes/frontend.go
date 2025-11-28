package routes

import (
	"github.com/labstack/echo/v4"
)

// AddFrontendRoutes registers route for folder with frontend.
func AddFrontendRoutes(e *echo.Echo) {
	e.Static("/frontend", "./frontend")
}
