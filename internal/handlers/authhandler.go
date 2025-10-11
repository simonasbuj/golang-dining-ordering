package handlers

import (
	"errors"
	"golang-dining-ordering/internal/customerrors"
	"golang-dining-ordering/internal/dto"
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

	var reqDto dto.SignUpRequestDto
	err := dto.Validate(c, &reqDto)
	if err != nil {
		h.logger.Error("failed to sign up user, request body validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "request body validation failed",
			"error":   err.Error(),
		})
	}

	newUserID, err := h.svc.CreateUser(c.Request().Context(), &reqDto)
	if err != nil {
		h.logger.Error("failed to create new user", "error", err)

		var uqConstraintErr *customerrors.UniqueConstraintError
		if errors.As(err, &uqConstraintErr) {
			return c.JSON(http.StatusConflict, map[string]string{
				"message": "user with this email already exists",
				"error":   err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "failed to create new user",
			"error":   err.Error(),
		})
	}

	h.logger.Info("new user created", "userID", newUserID)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "new user registered successfuly",
	})
}

func (h *AuthHandler) HandleSignIn(c echo.Context) error {
	h.logger.Info("handling signin")

	var reqDto dto.SignInRequestDto
	err := dto.Validate(c, &reqDto)
	if err != nil {
		h.logger.Error("failed to sign in user, request body validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "request body validation failed",
			"error":   err.Error(),
		})
	}

	resDto, err := h.svc.SignInUser(c.Request().Context(), &reqDto)
	if err != nil {
		h.logger.Error("failed to sign in user", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "failed to sign in user",
			"error":   err.Error(),
		})
	}

	h.logger.Info("signed in succesfully")
	return c.JSON(http.StatusOK, map[string]any{
		"message": "signed in successfuly",
		"data":    resDto,
	})
}
