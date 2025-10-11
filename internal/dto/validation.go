package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func Validate(ctx echo.Context, dto interface{}) error {
	if err := ctx.Bind(dto); err != nil {
		fmt.Print(err)
		return err
	}

	if err := validator.New().Struct(dto); err != nil {
		return err
	}

	return nil
}
