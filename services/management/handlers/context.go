package handlers

import (
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	restaurantIDParamName = "restaurant_id"
	menuItemIDParamName   = "item_id"
)

var errMissingUser = errors.New("missing user in context")

func getUserFromContext(c echo.Context) (*authDto.TokenClaimsDto, error) {
	user, ok := c.Get(middleware.ContextKeyAuthUser).(*authDto.TokenClaimsDto)
	if !ok || user.UserID == uuid.Nil {
		return nil, responses.JSONError(c, "unauthorized", errMissingUser)
	}

	return user, nil
}

func getUUUIDFromParams(c echo.Context, paramName string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(paramName))
	if err != nil {
		return uuid.Nil, responses.JSONError(
			c,
			"invalid id in url for "+paramName,
			fmt.Errorf("parsing uuid from params for %s: %w", paramName, err),
		)
	}

	return id, nil
}
