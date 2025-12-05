// Package responses provides helper functions for JSON API responses.
package responses

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a JSON success response.
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// JSONError sends a JSON error response with an optional status code.
func JSONError(c echo.Context, errMsg string, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	_ = c.JSON(statusCode, ErrorResponse{errMsg})

	return fmt.Errorf("errMsg: %w", err)
}

// JSONSuccess sends a JSON success response with optional data and status code.
func JSONSuccess(c echo.Context, message string, data any, status ...int) error {
	statusCode := http.StatusOK

	if len(status) > 0 {
		statusCode = status[0]
	}

	return c.JSON(statusCode, &SuccessResponse{
		Message: message,
		Data:    data,
	})
}
