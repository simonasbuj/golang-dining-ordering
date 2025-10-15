package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Validate binds and validates the given DTO from the request context.
func Validate(ctx echo.Context, dto interface{}) error {
	if err := ctx.Bind(dto); err != nil {
		return err
	}

	if err := validator.New().Struct(dto); err != nil {
		return err
	}

	return nil
}
