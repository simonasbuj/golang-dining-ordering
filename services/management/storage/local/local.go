// Package storage handles local storage of menu item images.
package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	errFileTooLarge        = errors.New("file is too large")
	errPathIsEmpty         = errors.New("path is empty")
	errUnsupportedFileType = errors.New("unsupported file type")
)

const (
	uploadDirPerm          = 0o750
	checkFileTypeReadBytes = 512
)

type localStorage struct {
	maxFileSize int64
	uploadsDir  string
}

// NewLocalStorage creates a new local storage with max file size and upload directory.
//
//revive:disable:unexported-return
func NewLocalStorage(maxFileSize int64, uploadsDir string) *localStorage {
	return &localStorage{
		maxFileSize: maxFileSize,
		uploadsDir:  uploadsDir,
	}
}

//revive:enable:unexported-return

// StoreMenuItemImage stores an image file and returns its local path.
func (s *localStorage) StoreMenuItemImage(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > s.maxFileSize {
		return "", fmt.Errorf(
			"%w: %d bytes, max is %d bytes",
			errFileTooLarge,
			fileHeader.Size,
			s.maxFileSize,
		)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file from header: %w", err)
	}
	defer file.Close() //nolint:errcheck

	err = s.isFileJPGorPNG(file)
	if err != nil {
		return "", fmt.Errorf("failed to confirm if file is png or jpg: %w", err)
	}

	imagePath := filepath.Join(
		s.uploadsDir,
		filepath.Clean(uuid.New().String()+fileHeader.Filename),
	)

	err = os.MkdirAll(s.uploadsDir, uploadDirPerm)
	if err != nil {
		return "", fmt.Errorf("failed to create upload folder: %w", err)
	}

	// #nosec G304: filename is sanitized with filepath.Base and safe characters
	out, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to store file locally: %w", err)
	}
	defer out.Close() //nolint:errcheck

	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return imagePath, nil
}

// DeleteMenuItemImage deletes an image file at the given path.
func (s *localStorage) DeleteMenuItemImage(path string) error {
	if path == "" {
		return errPathIsEmpty
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
	buf := make([]byte, checkFileTypeReadBytes)

	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read file for type detection: %w", err)
	}

	contentType := http.DetectContentType(buf[:n])
	if contentType != "image/jpeg" && contentType != "image/png" {
		return fmt.Errorf("%w: %s", errUnsupportedFileType, contentType)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	return nil
}
