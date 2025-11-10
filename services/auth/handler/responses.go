package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type errorResponse struct {
	Error string `json:"error"`
}

func jsonError(c echo.Context, errMsg string, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	_ = c.JSON(statusCode, errorResponse{errMsg})

	return fmt.Errorf("errMsg: %w", err)
}

func jsonSuccess(c echo.Context, message string, data any, status ...int) error {
	statusCode := http.StatusOK

	if len(status) > 0 {
		statusCode = status[0]
	}

	return c.JSON(statusCode, map[string]any{
		"message": message,
		"data":    data,
	})
}
