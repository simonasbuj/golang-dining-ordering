// Package dto contains data transfer objects for the orders service.
package dto

import "github.com/google/uuid"

// CurrentOrderDto represents the active order for a table.
type CurrentOrderDto struct {
	ID uuid.UUID `json:"id"`
}

// OrderItemRequestDto represents a request to add or delete an item from an order.
type OrderItemRequestDto struct {
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
	ItemID       uuid.UUID `json:"item_id"`
	Name         string    `json:"name"`
	PriceInCents int       `json:"price_in_cents"`
}
