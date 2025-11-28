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

	user, err := getUserFromContext(c)
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

// HandleGetRestaurants handles fetching a paginated list of restaurants.
func (h *RestaurantsHandler) HandleGetRestaurants(c echo.Context) error {
	var reqDto dto.GetRestaurantsReqDto

	err := validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	resDto, err := h.svc.GetRestaurants(c.Request().Context(), &reqDto)
	if err != nil {
		return responses.JSONError(c, "failed to fetch restaurants", err)
	}

	return responses.JSONSuccess(c, "restaurants fetched", resDto)
}

// HandleGetRestaurantByID handles fetching a single restaurant by its ID.
func (h *RestaurantsHandler) HandleGetRestaurantByID(c echo.Context) error {
	id, err := getUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	resDto, err := h.svc.GetRestaurantByID(c.Request().Context(), id)
	if err != nil {
		return responses.JSONError(c, "failed to fetch restaurant", err)
	}

	return responses.JSONSuccess(c, "restaurant fetched", resDto)
}

// HandleUpdateRestaurant updates a restaurant’s details.
func (h *RestaurantsHandler) HandleUpdateRestaurant(c echo.Context) error {
	id, err := getUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	user, err := getUserFromContext(c)
	if err != nil {
		return err
	}

	var reqDto dto.UpdateRestaurantRequestDto

	reqDto.ID = id
	reqDto.UserID = user.UserID

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	resDto, err := h.svc.UpdateRestaurant(c.Request().Context(), &reqDto)
	if err != nil {
		return responses.JSONError(c, "failed to update restaurant", err)
	}

	return responses.JSONSuccess(c, "restaurant updated", resDto)
}

// HandleCreateTable handles creating a new table for a restaurant.
func (h *RestaurantsHandler) HandleCreateTable(c echo.Context) error {
	id, err := getUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	user, err := getUserFromContext(c)
	if err != nil {
		return err
	}

	var reqDto dto.RestaurantTableDto

	reqDto.RestaurantID = id
	reqDto.UserID = user.UserID

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.CreateTable(c.Request().Context(), &reqDto)
	if err != nil {
		if errors.Is(err, services.ErrUserIsNotManager) {
			return responses.JSONError(
				c,
				"user is unauthorized to add tables for this restaurant",
				err,
				http.StatusUnauthorized,
			)
		}

		return responses.JSONError(c, "failed to add table", err)
	}

	return responses.JSONSuccess(c, "table added to restaurant", respDto)
}

// HandleGetTables fetches all tables belonging to a restaurant.
func (h *RestaurantsHandler) HandleGetTables(c echo.Context) error {
	id, err := getUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	respDto, err := h.svc.GetTables(c.Request().Context(), id)
	if err != nil {
		return responses.JSONError(c, "failed to fetch restaurant tables", err)
	}

	return responses.JSONSuccess(c, "tables fetched", respDto)
}
