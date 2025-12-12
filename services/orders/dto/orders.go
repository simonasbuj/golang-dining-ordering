// Package dto contains data transfer objects for the orders service.
package dto

import (
	"encoding/json"
	db "golang-dining-ordering/services/orders/db/generated"
	"time"

	"github.com/google/uuid"
)

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
	RestaurantID      uuid.UUID       `json:"restaurant_id"`
	RestaurantName    string          `json:"restaurant_name"`
	Status            db.OrderStatus  `json:"status"`
	Currency          string          `json:"currency"`
	TipAmountInCents  int             `json:"tip_amount_in_cents"`
	TotalPriceInCents int             `json:"total_price_in_cents"`
	UpdatedAt         time.Time       `json:"updated_at"`
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

// UpdateOrderReqDto represents a request payload to update order.
type UpdateOrderReqDto struct {
	OrderID          uuid.UUID       `json:"order_id"            validate:"required"`
	TipAmountInCents *int32          `json:"tip_amount_in_cents" validate:"omitempty,gte=0,lt=20000"`
	Status           *db.OrderStatus `json:"status"`
}

// RemoveWaiterReqDto represents request payload to unassing waiter from order.
type RemoveWaiterReqDto struct {
	ID uuid.UUID `json:"assign_id" validate:"required"`
}

// WSMessageType represents the type of a WebSocket message.
type WSMessageType string

const (
	// MsgUpdateOrder to update an order.
	MsgUpdateOrder WSMessageType = "update_order"
	// MsgAddItem to add an item to an order.
	MsgAddItem WSMessageType = "add_item"
	// MsgDeleteItem to delete an item from an order.
	MsgDeleteItem WSMessageType = "delete_item"
	// MsgError indicating an error.
	MsgError WSMessageType = "error"
)

// WSReqMessage is the envelope for messages received from the client.
type WSReqMessage struct {
	Type WSMessageType   `json:"type" validate:"required"`
	Data json.RawMessage `json:"data" validate:"required"`
}

// WSRespMessage is the envelope for messages sent back to the client.
type WSRespMessage struct {
	Type WSMessageType `json:"type"`
	Data any           `json:"data"`
}
