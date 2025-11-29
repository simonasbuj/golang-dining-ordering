// Package middleware is middleware
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
func AuthMiddleware(authServiceURL string, failOnMissingUser ...bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fail := true
			if len(failOnMissingUser) > 0 {
				fail = failOnMissingUser[0]
			}

			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return handleAuthError(
					c,
					fail,
					http.StatusUnauthorized,
					"missing Authorization header",
					nil,
					next,
				)
			}

			resp, err := callAuthService(c.Request().Context(), authServiceURL, token)
			if err != nil {
				return handleAuthError(
					c,
					fail,
					http.StatusInternalServerError,
					"failed to reach auth service",
					nil,
					next,
				)
			}
			defer resp.Body.Close() //nolint:errcheck

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)

				return handleAuthError(c, fail, resp.StatusCode, "unauthorized", body, next)
			}

			err = parseAndStoreAuthResponse(c, resp)
			if err != nil {
				return handleAuthError(
					c,
					fail,
					http.StatusInternalServerError,
					"failed to parse auth-service response",
					nil,
					next,
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
func callAuthService(ctx context.Context, url, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling auth service: %w", err)
	}

	return resp, nil
}

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

func handleAuthError(
	c echo.Context,
	fail bool,
	status int,
	message string,
	body []byte,
	next echo.HandlerFunc,
) error {
	if !fail {
		c.Set(ContextKeyAuthUser, &authDto.TokenClaimsDto{}) //nolint:exhaustruct

		return next(c)
	}

	if body != nil {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(status)
		_, _ = c.Response().Write(body)

		return errUnauthorized
	}

	err := c.JSON(status, map[string]string{"error": message})

	return fmt.Errorf("sending response to client: %w", err)
}
