package handlers

import (
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/services/management/dto"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getUserFromContext(c echo.Context) (*dto.TokenClaimsDto, error) {
	user, ok := c.Get("authUser").(*dto.TokenClaimsDto)
	if !ok || user.UserID == uuid.Nil {
		return nil, responses.JSONError(c, "unauthorized", errMissingUser)
	}

	return user, nil
}
