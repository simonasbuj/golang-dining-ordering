// Package handlers contains HTTP handlers for the application.
package handlers

import (
	"errors"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/services"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

var errMissingUser = errors.New("missing user in context")

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

	user, err := h.getUserFromContext(c)
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
	limitStr := c.QueryParam("limit")
	pageStr := c.QueryParam("page")

	limit := 10
	page := 1

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}

	reqDto := &dto.GetRestaurantsReqDto{
		Page:  page,
		Limit: limit,
	}

	resDto, err := h.svc.GetRestaurants(c.Request().Context(), reqDto)
	if err != nil {
		return responses.JSONError(c, "failed to fetch restaurants", err)
	}

	return responses.JSONSuccess(c, "restaurants fetched", resDto)
}

func (h *RestaurantsHandler) getUserFromContext(c echo.Context) (*dto.TokenClaimsDto, error) {
	user, ok := c.Get("authUser").(*dto.TokenClaimsDto)
	if !ok || user.UserID == "" {
		return nil, responses.JSONError(c, "unauthorized", errMissingUser)
	}

	return user, nil
}
