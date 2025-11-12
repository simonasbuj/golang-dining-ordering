package handlers

import (
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/services/management/services"

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
	user, err := getUserFromContext(c)
	if err != nil {
		return err
	}

	return responses.JSONSuccess(c, "gonna add new menu", user)
}
