// Package storage handles local storage of menu item images.
package storage

import (
	"context"
	"golang-dining-ordering/config"
	"golang-dining-ordering/services/management/storage/local"
	"golang-dining-ordering/services/management/storage/s3"
	"mime/multipart"
)

// Storage defines methods for storing and deleting menu item images.
type Storage interface {
	StoreMenuItemImage(ctx context.Context, fileHeader *multipart.FileHeader) (string, error)
	DeleteMenuItemImage(ctx context.Context, path string) error
}

// GetStorage returns the appropriate Storage implementation (S3 or local) based on storageType.
//
//nolint:ireturn
func GetStorage(
	ctx context.Context,
	storageType config.StorageType,
	cfg *config.AppConfig,
) Storage {
	switch storageType {
	case config.StorageTypeS3:
		return s3.NewS3Storage(
			ctx,
			cfg.S3Config.Key,
			cfg.S3Config.Secret,
			cfg.S3Config.URL,
			cfg.S3Config.Bucket,
		)
	case config.StorageTypeLocal:
		return local.NewLocalStorage(cfg.MaxImageSizeBytes, cfg.UploadsDirectory)
	default:
		return local.NewLocalStorage(cfg.MaxImageSizeBytes, cfg.UploadsDirectory)
	}
}
