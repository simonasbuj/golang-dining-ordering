// Package storage handles local storage of menu item images.
package storage

import (
	"context"
	"mime/multipart"
)

// Storage defines methods for storing and deleting menu item images.
type Storage interface {
	StoreMenuItemImage(ctx context.Context, fileHeader *multipart.FileHeader) (string, error)
	DeleteMenuItemImage(ctx context.Context, path string) error
}
