// Package dto contains data transfer objects for the orders service.
package dto

import "github.com/google/uuid"

// CurrentOrderDto represents the active order for a table.
type CurrentOrderDto struct {
	ID uuid.UUID
}
