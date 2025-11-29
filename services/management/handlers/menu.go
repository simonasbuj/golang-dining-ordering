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

// MenuHandler handles restaurant menu related HTTP requests.
type MenuHandler struct {
	svc services.MenuService
}

// NewMenuHandler creates a new RestauranMenuHandlertsHandler.
func NewMenuHandler(svc services.MenuService) *MenuHandler {
	return &MenuHandler{
		svc: svc,
	}
}

// HandleAddMenuCategory handles adding a new menu category.
func (h *MenuHandler) HandleAddMenuCategory(c echo.Context) error {
	restaurantID, err := GetUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	user, err := GetUserFromContext(c)
	if err != nil {
		return err
	}

	var reqDto dto.MenuCategoryDto

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	reqDto.RestaurantID = restaurantID

	resDto, err := h.svc.AddMenuCategory(c.Request().Context(), &reqDto, user)
	if err != nil {
		if errors.Is(err, services.ErrUserIsNotManager) {
			return responses.JSONError(
				c,
				"user is unauthorized to add menu items for this restaurant",
				err,
				http.StatusUnauthorized,
			)
		}

		return responses.JSONError(c, "failed to add menu category", err)
	}

	return responses.JSONSuccess(c, "menu category created", resDto)
}

// HandleAddMenuItem handles HTTP requests to add a new menu item.
func (h *MenuHandler) HandleAddMenuItem(c echo.Context) error {
	restaurantID, err := GetUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	user, err := GetUserFromContext(c)
	if err != nil {
		return err
	}

	var reqDto dto.MenuItemDto

	reqDto.RestaurantID = restaurantID

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	resDto, err := h.svc.AddMenuItem(c.Request().Context(), &reqDto, user)
	if err != nil {
		if errors.Is(err, services.ErrUserIsNotManager) {
			return responses.JSONError(
				c,
				"user is unauthorized to add menu items for this restaurant",
				err,
				http.StatusUnauthorized,
			)
		}

		return responses.JSONError(c, "failed to add menu item", err)
	}

	return responses.JSONSuccess(c, "new menu item added", resDto)
}

// HandleUpdateMenuItem updates a menu item for the specified restaurant.
func (h *MenuHandler) HandleUpdateMenuItem(c echo.Context) error {
	restaurantID, err := GetUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	itemID, err := GetUUUIDFromParams(c, menuItemIDParamName)
	if err != nil {
		return err
	}

	user, err := GetUserFromContext(c)
	if err != nil {
		return err
	}

	var reqDto dto.MenuItemDto

	reqDto.RestaurantID = restaurantID
	reqDto.ID = itemID

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.UpdateMenuItem(c.Request().Context(), &reqDto, user)
	if err != nil {
		if errors.Is(err, services.ErrUserIsNotManager) {
			return responses.JSONError(
				c,
				"user is unauthorized to edit this item",
				err,
				http.StatusUnauthorized,
			)
		}

		return responses.JSONError(c, "failed to add menu item", err)
	}

	return responses.JSONSuccess(c, "updated menu item", respDto)
}

// HandleGetMenuItems retrieves all menu categories and items for a restaurant.
func (h *MenuHandler) HandleGetMenuItems(c echo.Context) error {
	restaurantID, err := GetUUUIDFromParams(c, restaurantIDParamName)
	if err != nil {
		return err
	}

	resDto, err := h.svc.GetMenuItems(c.Request().Context(), restaurantID)
	if err != nil {
		return responses.JSONError(
			c,
			"failed to fetch menu items",
			err,
			http.StatusInternalServerError,
		)
	}

	return responses.JSONSuccess(c, "menu items fetched", resDto)
}
