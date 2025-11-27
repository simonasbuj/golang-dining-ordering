// Package validation provides helper functions for binding and validating
// request data transfer objects (DTOs) in Echo HTTP handlers.
package validation

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var errDtoMustBePointer = errors.New("ValidateDto requires a non-nil pointer")

// ValidateDto binds and validates the given DTO from the request context.
func ValidateDto(ctx echo.Context, dto any) error {
	v := reflect.ValueOf(dto)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("%w: got %T", errDtoMustBePointer, dto)
	}

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
