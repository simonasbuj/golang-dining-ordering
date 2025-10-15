package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Validate binds and validates the given DTO from the request context.
func Validate(ctx echo.Context, dto interface{}) error {
	err := ctx.Bind(dto)
	if err != nil {
		return fmt.Errorf("dto binding failed: %w", err)
	}

	err = validator.New().Struct(dto)
	if err != nil {
		return fmt.Errorf("dto validation failed: %w", err)
	}

	return nil
}
