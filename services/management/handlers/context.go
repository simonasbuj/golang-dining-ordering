package handlers

import (
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getUserFromContext(c echo.Context) (*authDto.TokenClaimsDto, error) {
	user, ok := c.Get("authUser").(*authDto.TokenClaimsDto)
	if !ok || user.UserID == uuid.Nil {
		return nil, responses.JSONError(c, "unauthorized", errMissingUser)
	}

	return user, nil
}
