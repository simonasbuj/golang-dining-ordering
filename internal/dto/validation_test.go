package dto_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang-dining-ordering/internal/dto"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestDto struct {
	Name     string `json:"name"     validate:"required"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func TestValidate_Success(t *testing.T) {
	t.Parallel()

	e := echo.New()
	inputDto := &TestDto{
		Name:     "sim",
		Email:    "sim@email.com",
		Password: "password123",
	}
	body, err := json.Marshal(inputDto)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	var payload TestDto
	err = dto.Validate(ctx, &payload)
	require.NoError(t, err)
	assert.Equal(t, inputDto, &payload)
}

func TestValidate_InvalidJsonBindError(t *testing.T) {
	t.Parallel()

	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid json}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	var payload TestDto
	err := dto.Validate(ctx, &payload)
	assert.Error(t, err)
}

func TestValidate_ValidationError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		name     string
		email    string
		password string
	}{
		{"name is missing", "", "sim@email.com", "password123"},
		{"email is not an email", "sim", "simnotanemail.com", "password123"},
		{"password is too short", "sim", "sim@email.com", "123"},
	}

	for _, testCase := range testCases {
		e := echo.New()

		inputDto := &TestDto{
			Name:     testCase.name,
			Email:    testCase.email,
			Password: testCase.password,
		}
		body, err := json.Marshal(inputDto)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		var payload TestDto
		err = dto.Validate(ctx, &payload)
		require.Error(t, err, "validation error should happen when %s", testCase.desc)

		var ve validator.ValidationErrors
		ok := errors.As(err, &ve)
		assert.True(t, ok, "error should be a validator.ValidationErrors when %s", testCase.desc)
	}
}
