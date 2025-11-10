// Package handler defines HTTP handlers for application endpoints.
package handler

import (
	"database/sql"
	"errors"
	"golang-dining-ordering/pkg/validation"
	"golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/auth/service"
	"log/slog"
	"net/http"
	"strings"

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
		return jsonError(c, err.Error(), err)
	}

	_, err = h.svc.SignUpUser(c.Request().Context(), &reqDto)
	if err != nil {
		return jsonError(c, "failed to register user", err)
	}

	return jsonSuccess(c, "new user registered successfully", nil, http.StatusCreated)
}

// HandleSignIn handles requests to sign in user.
func (h *Handler) HandleSignIn(c echo.Context) error {
	var reqDto dto.SignInRequestDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		return jsonError(c, err.Error(), err)
	}

	resDto, err := h.svc.SignInUser(c.Request().Context(), &reqDto)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ce.ErrUnauthorized) {
			return jsonError(c, "unauthorized", err, http.StatusUnauthorized)
		}

		return jsonError(c, "server error", err, http.StatusInternalServerError)
	}

	return jsonSuccess(c, "signed in successfully", resDto)
}

// HandleRefreshToken handles requests to refresh an authentication token.
func (h *Handler) HandleRefreshToken(c echo.Context) error {
	var reqDto dto.RefreshTokenRequestDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		return jsonError(c, err.Error(), err)
	}

	resDto, err := h.svc.RefreshToken(c.Request().Context(), reqDto.RefreshToken)
	if err != nil {
		if errors.Is(err, ce.ErrMissingClaims) || errors.Is(err, ce.ErrInvalidTokenData) ||
			errors.Is(
				err,
				ce.ErrParseToken,
			) || errors.Is(err, ce.ErrInvalidTokenVersion) || errors.Is(err, ce.ErrInvalidTokenType) {
			return jsonError(c, "invalid token data", err)
		}

		return jsonError(c, "failed to refresh token", err, http.StatusInternalServerError)
	}

	return jsonSuccess(c, "refreshed token successfully", resDto)
}

// HandleLogout handles requests to logout user by increastin token version in database.
func (h *Handler) HandleLogout(c echo.Context) error {
	var reqDto dto.LogoutRequestDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		return jsonError(c, err.Error(), err)
	}

	err = h.svc.LogoutUser(c.Request().Context(), reqDto.Token)
	if err != nil {
		if errors.Is(err, ce.ErrMissingClaims) || errors.Is(err, ce.ErrInvalidTokenData) ||
			errors.Is(
				err,
				ce.ErrParseToken,
			) || errors.Is(err, ce.ErrInvalidTokenVersion) || errors.Is(err, ce.ErrInvalidTokenType) {
			return jsonError(c, "invalid token data", err)
		}

		return jsonError(c, "logout failed", err, http.StatusInternalServerError)
	}

	return jsonSuccess(c, "logged out successfully", nil)
}

// HandleAuthorize validates the token from the request.
func (h *Handler) HandleAuthorize(c echo.Context) error {
	tokenStr := c.Request().Header.Get("Authorization")
	if tokenStr == "" {
		return jsonError(
			c,
			"missing token",
			ce.ErrMissingToken,
			http.StatusUnauthorized,
		)
	}

	trimmedToken := strings.TrimPrefix(tokenStr, "Bearer ")

	resDto, err := h.svc.Authorize(c.Request().Context(), trimmedToken)
	if errors.Is(err, ce.ErrMissingClaims) || errors.Is(err, ce.ErrInvalidTokenData) ||
		errors.Is(
			err,
			ce.ErrParseToken,
		) || errors.Is(err, ce.ErrInvalidTokenVersion) || errors.Is(err, ce.ErrInvalidTokenType) {
		return jsonError(c, "invalid token data", err)
	}

	return jsonSuccess(c, "authorized", resDto)
}
