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
	restaurantID := c.Param("restaurant_id")

	user, err := getUserFromContext(c)
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
