// Package handler defines HTTP handlers for application endpoints.
package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/validation"
	"golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/auth/service"
	"log/slog"
	"net/http"

	ce "golang-dining-ordering/services/auth/customerrors"

	"github.com/labstack/echo/v4"
)

// Handler handles authentication-related HTTP requests.
type Handler struct {
	logger *slog.Logger
	svc    service.Service
}

// NewAuthHandler creates a new AuthHandler for handling authentication requests.
func NewAuthHandler(logger *slog.Logger, svc service.Service) *Handler {
	return &Handler{
		logger: logger,
		svc:    svc,
	}
}

// HandleSignUp handles requests to sign up user.
func (h *Handler) HandleSignUp(c echo.Context) error {
	var reqDto dto.SignUpRequestDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "request body validation failed", err.Error())

		return fmt.Errorf("request body validation failed: %w", err)
	}

	_, err = h.svc.SignUpUser(c.Request().Context(), &reqDto)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "failed to register user", "invalid request")

		return fmt.Errorf("failed to create new user: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "new user registered successfully",
	})
}

// HandleSignIn handles requests to sign in user.
func (h *Handler) HandleSignIn(c echo.Context) error {
	var reqDto dto.SignInRequestDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "request body validation failed", err.Error())

		return fmt.Errorf("request body validation failed: %w", err)
	}

	resDto, err := h.svc.SignInUser(c.Request().Context(), &reqDto)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ce.ErrUnauthorized) {
			jsonError(c, http.StatusUnauthorized, "failed to sign in user", "unauthorized")

			return fmt.Errorf("failed to sign in user: %w", err)
		}

		jsonError(c, http.StatusInternalServerError, "failed to sign in user", err.Error())

		return fmt.Errorf("failed to sign in user: %w", err)
	}

	return jsonSuccess(c, http.StatusOK, "signed in successfully", resDto)
}

// HandleRefreshToken handles requests to refresh an authentication token.
func (h *Handler) HandleRefreshToken(c echo.Context) error {
	var reqDto dto.RefreshTokenRequestDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		jsonError(c, http.StatusBadRequest, "request body validation failed", err.Error())

		return fmt.Errorf("request body validation failed: %w", err)
	}

	resDto, err := h.svc.RefreshToken(c.Request().Context(), reqDto.RefreshToken)
	if err != nil {
		if errors.Is(err, ce.ErrMissingClaims) {
			
		}
		jsonError(c, http.StatusInternalServerError, "failed to refresh token", "processing error")

		return fmt.Errorf("failed to refresh token: %w", err)
	}

	return jsonSuccess(c, http.StatusOK, "signed in successfully", resDto)
}
