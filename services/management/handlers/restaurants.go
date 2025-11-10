package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
)


type RestaurantsHandler struct {}

func NewRestaurantsHandler() *RestaurantsHandler {
	return &RestaurantsHandler{}
}

func (h *RestaurantsHandler) HandleCreateRestaurant(c echo.Context) error {
	fmt.Println("this gon be restaurant creation")
	return nil
}