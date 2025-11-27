package handlers

import (
	"errors"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/middleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	errMissingUser           = errors.New("missing user in context")
	errIncorrectRestaurantID = errors.New("incorrect restaurant id in url")
)

func getUserFromContext(c echo.Context) (*authDto.TokenClaimsDto, error) {
	user, ok := c.Get(middleware.ContextKeyAuthUser).(*authDto.TokenClaimsDto)
	if !ok || user.UserID == uuid.Nil {
		return nil, responses.JSONError(c, "unauthorized", errMissingUser)
	}

	return user, nil
}

func getRestaurantFromParams(c echo.Context) (uuid.UUID, error) {
	restaurantID, err := uuid.Parse(c.Param("restaurant_id"))
	if err != nil {
		return uuid.Nil, responses.JSONError(
			c,
			errIncorrectRestaurantID.Error(),
			errIncorrectRestaurantID,
		)
	}

	return restaurantID, nil
}
