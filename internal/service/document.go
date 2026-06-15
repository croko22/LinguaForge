package service

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MaxUploadSize is the maximum allowed file size for uploads (50 MB).
const MaxUploadSize int64 = 50 * 1024 * 1024

// Sentinel errors for document operations.
var (
	ErrInvalidFileType = fmt.Errorf("invalid file type: not supported by any parser")
	ErrFileTooLarge    = fmt.Errorf("file too large: maximum size is 50 MB")
	ErrNoCover         = fmt.Errorf("no cover available")
)

// fileTypeFromFilename extracts the file extension without the leading dot.
func fileTypeFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimPrefix(strings.ToLower(ext), ".")
}

// roughTokenCount estimates the number of tokens in a text string.
func roughTokenCount(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Fields(content))
}
