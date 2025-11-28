// Package middleware is middleware
package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	"io"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
)

var errUnauthorized = errors.New("response from auth-service: unauthorized")

// ContextKeyAuthUser is the key used to store the authenticated user in echo.Context.
const ContextKeyAuthUser = "authUser"

// AuthResponse is response body from auth-service.
type AuthResponse struct {
	Data authDto.TokenClaimsDto `json:"data"`
}

// AuthMiddleware validates JWT tokens by delegating to the Auth service.
func AuthMiddleware(authServiceURL string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing Authorization header",
				})
			}

			req, err := http.NewRequestWithContext(
				c.Request().Context(),
				http.MethodPost,
				authServiceURL,
				nil,
			)
			if err != nil {
				return c.JSON(
					http.StatusInternalServerError,
					map[string]string{"error": "failed to create auth request"},
				)
			}

			req.Header.Set("Authorization", token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return c.JSON(
					http.StatusInternalServerError,
					map[string]string{"error": "failed to reach auth service"},
				)
			}
			defer resp.Body.Close() //nolint:errcheck

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)

				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				c.Response().WriteHeader(resp.StatusCode)
				_, _ = c.Response().Write(body)

				return errUnauthorized
			}

			err = parseAndStoreAuthResponse(c, resp)
			if err != nil {
				return responses.JSONError(
					c,
					"failed to parse auth-service response",
					err,
					http.StatusInternalServerError,
				)
			}

			return next(c)
		}
	}
}

// RoleMiddleware returns an Echo middleware that allows access only to users
// with one of the specified roles. It reads the authenticated user from the context.
func RoleMiddleware(allowedRoles ...authDto.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rawClaims := c.Get(ContextKeyAuthUser)
			if rawClaims == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "missing claims")
			}

			claims, ok := rawClaims.(*authDto.TokenClaimsDto)
			if !ok || claims == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "failed to parse claims into dto")
			}

			if slices.Contains(allowedRoles, claims.Role) {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
		}
	}
}

// parseAndStoreAuthResponse parses the auth service response and stores the claims in context.
func parseAndStoreAuthResponse(c echo.Context, resp *http.Response) error {
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response body: %w", err)
	}

	var authResp AuthResponse

	err = json.Unmarshal(body, &authResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w", err)
	}

	c.Set(ContextKeyAuthUser, &authResp.Data)

	return nil
}
