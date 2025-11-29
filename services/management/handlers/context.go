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

// GetUserFromContext parses User dto from echo context.
func GetUserFromContext(
	c echo.Context,
	failOnMissingUser ...bool,
) (*authDto.TokenClaimsDto, error) {
	fail := true
	if len(failOnMissingUser) > 0 {
		fail = failOnMissingUser[0]
	}

	user, ok := c.Get(middleware.ContextKeyAuthUser).(*authDto.TokenClaimsDto)
	if (!ok || user.UserID == uuid.Nil) && fail {
		return nil, responses.JSONError(c, "unauthorized", errMissingUser)
	}

	return user, nil
}

// GetUUUIDFromParams parses UUID from provided url param.
func GetUUUIDFromParams(c echo.Context, paramName string) (uuid.UUID, error) {
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
