// Package dto contains data transfer objects for the orders service.
package dto

import "github.com/google/uuid"

type CurrentOrderDto struct {
	ID uuid.UUID
}
