package responses

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

var errTestError = errors.New("underlying error")

func TestJSONError(t *testing.T) {
	t.Parallel()

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	retErr := JSONError(c, "something went wrong", errTestError)
	require.Error(t, retErr)
	require.Contains(t, retErr.Error(), "errMsg")

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "something went wrong", resp.Error)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestJSONError_WithCustomStatus(t *testing.T) {
	t.Parallel()

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	retErr := JSONError(c, "oops", errTestError, http.StatusTeapot)
	require.Error(t, retErr)
	require.Equal(t, http.StatusTeapot, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "oops", resp.Error)
}

func TestJSONSuccess(t *testing.T) {
	t.Parallel()

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	data := map[string]interface{}{"anything": "something"}
	retErr := JSONSuccess(c, "all good", data)
	require.NoError(t, retErr)

	var resp SuccessResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "all good", resp.Message)
	require.Equal(t, data, resp.Data)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestJSONSuccess_WithCustomStatus(t *testing.T) {
	t.Parallel()

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)

	retErr := JSONSuccess(c, "done", nil, http.StatusCreated)
	require.NoError(t, retErr)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp SuccessResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, "done", resp.Message)
	require.Nil(t, resp.Data)
}
