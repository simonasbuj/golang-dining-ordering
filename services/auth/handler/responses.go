package handler

import "github.com/labstack/echo/v4"

type errorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func jsonError(c echo.Context, status int, msg, errMsg string) {
	_ = c.JSON(status, errorResponse{msg, errMsg})
}

func jsonSuccess(c echo.Context, status int, message string, data any) error {
	return c.JSON(status, map[string]any{
		"message": message,
		"data":    data,
	})
}
