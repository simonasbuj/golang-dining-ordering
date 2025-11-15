package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type localStorage struct {
	maxFileSize int64
}

func NewLocalStorage(maxFileSize int64) *localStorage {
	return &localStorage{
		maxFileSize: maxFileSize,
	}
}

func (s *localStorage) StoreMenuItemImage(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > s.maxFileSize {
        return "", fmt.Errorf("file is too large: %d bytes, max is %d bytes", fileHeader.Size, s.maxFileSize)
    }
	
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file from header: %w", err)
	}
	defer file.Close()

	err = s.isFileJPGorPNG(file)
	if err != nil {
		return "", fmt.Errorf("failed to confirm if file is png or jpg: %w", err)
	}

	imagePath := "uploads/images/" +  uuid.New().String() + fileHeader.Filename
	out, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to store file locally: %w", err)
	}
	defer out.Close()
	io.Copy(out, file)

	return imagePath, nil
}

func (s *localStorage) DeleteMenuItemImage(path string) error {
	fmt.Println("TYRING TO DELETE: ", path)
	if path == "" {
        return fmt.Errorf("path is empty")
    }

    err := os.Remove(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return fmt.Errorf("failed to delete file: %w", err)
    }

    return nil
}

func (s *localStorage) isFileJPGorPNG(file multipart.File) error {
    buf := make([]byte, 512)
    n, err := file.Read(buf)
    if err != nil && err != io.EOF {
        return fmt.Errorf("failed to read file for type detection: %w", err)
    }

    contentType := http.DetectContentType(buf[:n])
    if contentType != "image/jpeg" && contentType != "image/png" {
        return fmt.Errorf("unsupported file type: %s", contentType)
    }

    if _, err := file.Seek(0, io.SeekStart); err != nil {
        return fmt.Errorf("failed to reset file pointer: %w", err)
    }

    return nil
}
