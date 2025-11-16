package handlers

import (
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/services"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var (
	errInvalidPriceValue       = errors.New("invalid price field value")
	errInvalidIsAvailableValue = errors.New("invalid is_available field value")
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

// HandleAddMenuItem handles HTTP requests to add a new menu item.
func (h *MenuHandler) HandleAddMenuItem(c echo.Context) error {
	user, err := getUserFromContext(c)
	if err != nil {
		return err
	}

	reqDto, err := h.getItemFormFields(c)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	resDto, err := h.svc.AddMenuItem(c.Request().Context(), reqDto, user)
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

func (h *MenuHandler) getItemFormFields(c echo.Context) (*dto.MenuItemDto, error) {
	var reqDto dto.MenuItemDto

	reqDto.RestaurantID = c.Param("restaurant_id")
	reqDto.CategoryID = c.FormValue("category_id")
	reqDto.Name = c.FormValue("name")
	reqDto.Description = c.FormValue("description")

	priceStr := c.FormValue("price")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, errInvalidPriceValue
	}

	reqDto.Price = price

	isAvailableStr := c.FormValue("is_available")
	if isAvailableStr != "" {
		isAvailable, err := strconv.ParseBool(isAvailableStr)
		if err != nil {
			return nil, errInvalidIsAvailableValue
		}

		reqDto.IsAvailable = isAvailable
	}

	reqDto.FileHeader, _ = c.FormFile("image")

	err = validator.New().Struct(reqDto)
	if err != nil {
		return nil, fmt.Errorf("input form validation failed: %w", err)
	}

	return &reqDto, nil
}
