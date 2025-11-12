// Package handlers contains HTTP handlers for the application.
package handlers

import (
	"errors"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/services"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RestaurantsHandler handles restaurant-related HTTP requests.
type RestaurantsHandler struct {
	svc services.RestaurantService
}

// NewRestaurantsHandler creates a new RestaurantsHandler.
func NewRestaurantsHandler(svc services.RestaurantService) *RestaurantsHandler {
	return &RestaurantsHandler{
		svc: svc,
	}
}

// HandleCreateRestaurant handles creating a new restaurant.
func (h *RestaurantsHandler) HandleCreateRestaurant(c echo.Context) error {
	var reqDto dto.CreateRestaurantDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	user, err := h.getUserID(c)
	if err != nil {
		return err
	}
	reqDto.UserID = user.UserID

	resDto, err := h.svc.CreateRestaurant(c.Request().Context(), &reqDto)
	if err != nil {
		return responses.JSONError(
			c,
			"failed to create new restaurant",
			err,
			http.StatusInternalServerError,
		)
	}

	return responses.JSONSuccess(c, "new restaurant created", resDto)
}

func (h *RestaurantsHandler) getUserID(c echo.Context) (*dto.TokenClaimsDto, error) {
	user, ok := c.Get("authUser").(*dto.TokenClaimsDto)
	if !ok || user.UserID == "" {
		return nil, responses.JSONError(c, "unauthorized", errors.New("missing user ID in context"))
	}

	return user, nil
}
