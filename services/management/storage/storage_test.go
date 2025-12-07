package storage

import (
	"context"
	"golang-dining-ordering/config"
	"golang-dining-ordering/services/management/storage/local"
	"golang-dining-ordering/services/management/storage/s3"
	"reflect"
	"testing"
)

func TestGetStorage(t *testing.T) {
	t.Parallel()

	cfg := &config.AppConfig{ //nolint:exhaustruct
		MaxImageSizeBytes: 1024 * 1024,
		UploadsDirectory:  "/tmp/uploads",
		S3Config: config.S3Config{
			Key:    "key",
			Secret: "secret",
			URL:    "https://s3.example.com",
			Bucket: "bucket",
		},
	}

	tests := []struct {
		name        string
		storageType config.StorageType
		wantType    reflect.Type
	}{
		{
			name:        "S3 storage",
			storageType: config.StorageTypeS3,
			wantType: reflect.TypeOf(
				s3.NewS3Storage(
					context.Background(),
					cfg.S3Config.Key,
					cfg.S3Config.Secret,
					cfg.S3Config.URL,
					cfg.S3Config.Bucket,
				),
			),
		},
		{
			name:        "Local storage",
			storageType: config.StorageTypeLocal,
			wantType: reflect.TypeOf(
				local.NewLocalStorage(cfg.MaxImageSizeBytes, cfg.UploadsDirectory),
			),
		},
		{
			name:        "Default storage",
			storageType: "unknown",
			wantType: reflect.TypeOf(
				local.NewLocalStorage(cfg.MaxImageSizeBytes, cfg.UploadsDirectory),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetStorage(context.Background(), tt.storageType, cfg)
			if reflect.TypeOf(got) != tt.wantType {
				t.Errorf("GetStorage() = %T, want %v", got, tt.wantType)
			}
		})
	}
}
