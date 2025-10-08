package handlers

import (
	"golang-dining-ordering/internal/services"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	logger *slog.Logger
	svc    services.UserService
}

func NewAuthHandler(logger *slog.Logger, svc services.UserService) *AuthHandler {
	return &AuthHandler{
		logger: logger,
		svc:    svc,
	}
}

func (h *AuthHandler) HandleSignUp(c echo.Context) error {
	h.logger.Info("handling signup")

	user, err := h.svc.CreateUser(c.Request().Context())
	if err != nil {
		h.logger.Error("failed to create new user", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "failed to create new user",
			"error":   err.Error(),
		})
	}

	h.logger.Info("new user created", "user", user)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "new user created",
		"user":    user.ID,
	})

}
