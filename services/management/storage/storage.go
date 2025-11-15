package storage

import "mime/multipart"

type Storage interface {
	StoreMenuItemImage(fileHeader *multipart.FileHeader) (string, error) 
}