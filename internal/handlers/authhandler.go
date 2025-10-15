// Package handlers defines HTTP handlers for application endpoints.
package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	ce "golang-dining-ordering/internal/customerrors"
	"golang-dining-ordering/internal/dto"
	"golang-dining-ordering/internal/services"

	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	logger *slog.Logger
	svc    services.AuthService
}

// NewAuthHandler creates a new AuthHandler for handling authentication requests.
func NewAuthHandler(logger *slog.Logger, svc services.AuthService) *AuthHandler {
	return &AuthHandler{
		logger: logger,
		svc:    svc,
	}
}

// HandleSignUp handles requests to sign up user.
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

	newUserID, err := h.svc.SignUpUser(c.Request().Context(), &reqDto)
	if err != nil {
		h.logger.Error("failed to create new user", "error", err)

		// using custom error to figure out which http status to send back to the client
		var uqConstraintErr *ce.UniqueConstraintError
		if errors.As(err, &uqConstraintErr) {
			return c.JSON(http.StatusConflict, map[string]string{
				"message": "user with this email already exists",
				"error":   err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "failed to create new user",
			"error":   "hi",
		})
	}

	h.logger.Info("new user created", "userID", newUserID)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "new user registered successfully",
	})
}

// HandleSignIn handles requests to sign in user.
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

		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ce.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "failed to sign in user",
				"error":   "unauthorized",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "failed to sign in user",
			"error":   err.Error(),
		})
	}

	h.logger.Info("signed in successfully")

	return c.JSON(http.StatusOK, map[string]any{
		"message": "signed in successfully",
		"data":    resDto,
	})
}

// HandleRefreshToken handles requests to refresh an authentication token.
func (h *AuthHandler) HandleRefreshToken(c echo.Context) error {
	h.logger.Info("handling refresh token")

	var reqDto dto.RefreshTokenRequestDto

	err := dto.Validate(c, &reqDto)
	if err != nil {
		h.logger.Error("failed to refresh token, request body validation failed", "error", err)

		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "request body validation failed",
			"error":   err.Error(),
		})
	}

	resDto, err := h.svc.RefreshToken(c.Request().Context(), reqDto.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "failed to refresh token",
			"error":   err.Error(),
		})
	}

	h.logger.Info("refreshed token successfully")

	return c.JSON(http.StatusOK, map[string]any{
		"message": "signed in successfully",
		"data":    resDto,
	})
}
