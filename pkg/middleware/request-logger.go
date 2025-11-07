// Package middleware contains Echo middlewares for logging and request handling.
package middleware

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestLogger logs start, end, and errors for each route.
func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()
			corrID := uuid.New().String()

			logger.Info("starting request",
				"method", req.Method,
				"path", req.URL.Path,
				"correlation_id", corrID,
			)

			err := next(c)

			duration := time.Since(start)

			if err != nil {
				c.Error(err)
				logger.Error("error while handling request",
					"method", req.Method,
					"path", req.URL.Path,
					"duration", duration,
					"status", res.Status,
					"error", err,
					"correlation_id", corrID,
				)
			} else {
				logger.Info("finished request",
					"method", req.Method,
					"path", req.URL.Path,
					"duration", duration,
					"status", res.Status,
					"correlation_id", corrID,
				)
			}

			return err
		}
	}
}
