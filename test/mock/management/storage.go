package management

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
)

var errStorageFailed = errors.New("storage failed")

type mockStorage struct{}

// NewMockStorage creates new mock storage.
func NewMockStorage() *mockStorage { //nolint:revive
	return &mockStorage{}
}

func (*mockStorage) StoreMenuItemImage(
	_ context.Context,
	fh *multipart.FileHeader,
) (string, error) {
	name := strings.ToLower(fh.Filename)

	if !strings.HasSuffix(name, ".jpg") &&
		!strings.HasSuffix(name, ".jpeg") &&
		!strings.HasSuffix(name, ".png") {
		return "", errStorageFailed
	}

	return testItemImagePath, nil
}

func (*mockStorage) DeleteMenuItemImage(_ context.Context, _ string) error {
	return nil
}
