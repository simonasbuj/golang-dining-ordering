package dto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type TestDto struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func TestValidate_Success(t *testing.T) {
	e := echo.New()
	dto := &TestDto{
		Name:     "sim sim",
		Email:    "sim@email.com",
		Password: "password123",
	}
	body, _ := json.Marshal(dto)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	var payload TestDto
	err := Validate(ctx, &payload)
	assert.NoError(t, err)
	assert.Equal(t, dto, &payload)
}

func TestValidate_InvalidJsonBindError(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid json}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	var payload TestDto
	err := Validate(ctx, &payload)
	assert.Error(t, err)
}

func TestValidate_ValidationError(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		email    string
		password string
	}{
		{"name is missing", "", "sim@email.com", "password123"},
		{"email is not an email", "sim sim", "simnotanemail.com", "password123"},
		{"password is too short", "sim sim", "sim@email.com", "123"},
	}

	for _, tc := range testCases {
		e := echo.New()

		dto := &TestDto{
			Name:     tc.name,
			Email:    tc.email,
			Password: tc.password,
		}
		body, _ := json.Marshal(dto)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		var payload TestDto
		err := Validate(ctx, &payload)
		assert.Error(t, err, fmt.Sprintf("validation error should happen when %s", tc.desc))
		_, ok := err.(validator.ValidationErrors)
		assert.True(t, ok, fmt.Sprintf("error should be a validator.ValidationErrors when %s", tc.desc))
	}
}
