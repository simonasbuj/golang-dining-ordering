package local

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeMultipartFile struct {
	*bytes.Reader
}

func (f *fakeMultipartFile) Close() error {
	return nil
}

func padBytes(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}

	padded := make([]byte, size)
	copy(padded, b)

	return padded
}

func TestIsFileJPGorPNG(t *testing.T) {
	t.Parallel()

	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}

	tests := []struct {
		name      string
		file      multipart.File
		wantError bool
	}{
		{
			name: "JPEG file",
			file: &fakeMultipartFile{
				bytes.NewReader(padBytes(jpegHeader, checkFileTypeReadBytes)),
			},
			wantError: false,
		},
		{
			name: "PNG file",
			file: &fakeMultipartFile{
				bytes.NewReader(padBytes(pngHeader, checkFileTypeReadBytes)),
			},
			wantError: false,
		},
		{
			name: "unsupported file type",
			file: &fakeMultipartFile{
				bytes.NewReader(padBytes([]byte("hello world"), checkFileTypeReadBytes)),
			},
			wantError: true,
		},
	}

	s := NewLocalStorage(80000, "uploads")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := s.isFileJPGorPNG(tt.file)
			if (err != nil) != tt.wantError {
				t.Errorf("expected error: %v, got: %v", tt.wantError, err)
			}
		})
	}
}

func TestDeleteMenuItemImage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	s := NewLocalStorage(1024*1024, tmpDir)

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()

		err := s.DeleteMenuItemImage(context.Background(), "")
		if !errors.Is(err, errPathIsEmpty) {
			t.Errorf("expected errPathIsEmpty, got %v", err)
		}
	})

	t.Run("file exists", func(t *testing.T) {
		t.Parallel()

		tmpFile, err := os.CreateTemp(t.TempDir(), "testfile-*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		err = tmpFile.Close()
		require.NoError(t, err)

		err = s.DeleteMenuItemImage(context.Background(), tmpFile.Name())
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		_, err = os.Stat(tmpFile.Name())
		if !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted, but it still exists")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		t.Parallel()

		path := "non-existent-file.txt"

		err := s.DeleteMenuItemImage(context.Background(), path)
		if err != nil {
			t.Errorf("expected nil error for non-existent file, got %v", err)
		}
	})
}

func createMultipartFile(
	t *testing.T,
	fieldName, filename string,
	content []byte,
) (*multipart.FileHeader, int64) {
	t.Helper()

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = fw.Write(content)
	if err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	err = w.Close()
	require.NoError(t, err)

	// create a new http request and parse it as multipart form
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	err = req.ParseMultipartForm(int64(len(content)))
	if err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}

	fileHeaders := req.MultipartForm.File[fieldName]
	if len(fileHeaders) == 0 {
		t.Fatalf("no files found in multipart form")
	}

	return fileHeaders[0], int64(len(content))
}

func TestStoreMenuItemImage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	s := NewLocalStorage(1024*1024, tmpDir) // max 1mb

	tests := []struct {
		name      string
		filename  string
		content   []byte
		wantError bool
	}{
		{
			name:     "valid PNG",
			filename: "test.png",
			content: append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
				bytes.Repeat([]byte{0}, 512)...),
			wantError: false,
		},
		{
			name:     "valid JPEG",
			filename: "test.jpg",
			content: append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
				bytes.Repeat([]byte{0}, 512)...),
			wantError: false,
		},
		{
			name:      "unsupported file type",
			filename:  "test.txt",
			content:   []byte("hello world"),
			wantError: true,
		},
		{
			name:      "file too large",
			filename:  "large.png",
			content:   bytes.Repeat([]byte{0x89, 0x50, 0x4E, 0x47}, 1024*1024),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fileHeader, size := createMultipartFile(t, "file", tt.filename, tt.content)
			fileHeader.Size = size

			path, err := s.StoreMenuItemImage(context.Background(), fileHeader)
			if (err != nil) != tt.wantError {
				t.Errorf("expected error: %v, got: %v", tt.wantError, err)
			}

			if !tt.wantError {
				_, err := os.Stat(path)
				if err != nil {
					t.Errorf("expected file to exist at %s, got error: %v", path, err)
				}
			}
		})
	}
}
