// Package middleware is middleware
package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/services/management/dto"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// AuthResponse is response body from auth-service.
type AuthResponse struct {
	Data dto.TokenClaimsDto `json:"data"`
}

// AuthMiddleware validates JWT tokens by delegating to the Auth service.
func AuthMiddleware(authServiceURL string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Println("MIDDLEWARE HIT")

			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing Authorization header",
				})
			}

			req, err := http.NewRequest(
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
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)

				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				c.Response().WriteHeader(resp.StatusCode)
				_, _ = c.Response().Write(body)

				return errors.New("response from auth-service: unauthorized")
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

// parseAndStoreAuthResponse parses the auth service response and stores the claims in context.
func parseAndStoreAuthResponse(c echo.Context, resp *http.Response) error {
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response body: %w", err)
	}

	// Decode JSON into struct
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Store in context
	c.Set("authUser", &authResp.Data)

	return nil
}
