// Package storage handles local storage of menu item images.
package storage

import "mime/multipart"

// Storage defines methods for storing and deleting menu item images.
type Storage interface {
	StoreMenuItemImage(fileHeader *multipart.FileHeader) (string, error)
	DeleteMenuItemImage(path string) error
}
