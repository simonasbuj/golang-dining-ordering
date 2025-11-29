// Package dto contains data transfer objects for the orders service.
package dto

import "github.com/google/uuid"

// CurrentOrderDto represents the active order for a table.
type CurrentOrderDto struct {
	ID uuid.UUID
}

// AddItemToOrderRequestDto represents a request to add an item to an order.
type AddItemToOrderRequestDto struct {
	ItemID uuid.UUID `json:"item_id" validate:"required"`
}

// OrderDto represents a full order with items and totals.
type OrderDto struct {
	ID                uuid.UUID       `json:"id"`
	RestaurantID      uuid.UUID       `json:"-"`
	Status            string          `json:"status"`
	Currency          string          `json:"currency"`
	TipAmountInCents  int             `json:"tip_amount_in_cents"`
	TotalPriceInCents int             `json:"total_price_in_cents"`
	Items             []*OrderItemDto `json:"items"`
}

// OrderItemDto represents a single item within an order.
type OrderItemDto struct {
	ID           uuid.UUID `json:"id"`
	RestaurantID uuid.UUID `json:"-"`
	Name         string    `json:"name"`
	PriceInCents int       `json:"price_in_cents"`
}
