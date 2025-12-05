package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

var errHandlerError = errors.New("handler failed")

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	e := echo.New()

	logger := slog.New(slog.DiscardHandler)

	nextHandler := func(_ echo.Context) error {
		return nil
	}

	h := RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_WithError(t *testing.T) {
	t.Parallel()

	e := echo.New()
	logger := slog.New(slog.DiscardHandler)

	nextHandler := func(_ echo.Context) error {
		return errHandlerError
	}

	h := RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h(c)
	require.Error(t, err)
	require.Equal(t, errHandlerError, err)
}
