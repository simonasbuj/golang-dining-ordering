package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type localStorage struct {}

func NewLocalStorage() *localStorage {
	return &localStorage{}
}

func (s *localStorage) StoreMenuItemImage(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file from header: %w", err)
	}
	defer file.Close()

	imagePath := "uploads/images/" + fileHeader.Filename
	out, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to store file locally: %w", err)
	}
	defer out.Close()
	io.Copy(out, file)

	return imagePath, nil
}