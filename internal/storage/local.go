package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
)

// LocalFileStorage stores files on the local filesystem.
type LocalFileStorage struct {
	baseDir string
}

// NewLocalFileStorage creates a LocalFileStorage and ensures the base
// directory exists.
func NewLocalFileStorage(baseDir string) (*LocalFileStorage, error) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolve base dir: %w", err)
	}
	if err := os.MkdirAll(absDir, 0750); err != nil {
		return nil, fmt.Errorf("create base dir: %w", err)
	}
	return &LocalFileStorage{baseDir: absDir}, nil
}

// Store saves a file from reader and returns the relative storage path.
// The stored filename is {uuid}_{original} to avoid collisions.
func (s *LocalFileStorage) Store(_ context.Context, filename string, reader io.Reader) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}

	relPath := fmt.Sprintf("%s_%s", id.String(), path.Clean(filename))
	fullPath := filepath.Join(s.baseDir, relPath)

	// Prevent path traversal even on the relative portion.
	if !isWithin(s.baseDir, fullPath) {
		return "", fmt.Errorf("resolved path escapes base directory")
	}

	f, err := os.Create(fullPath) // #nosec G304 — validated by isWithin check above
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	if written == 0 {
		return "", fmt.Errorf("no bytes written to file")
	}

	return relPath, nil
}

// Get opens a file for reading by its relative storage path.
func (s *LocalFileStorage) Get(_ context.Context, pathStr string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, path.Clean(pathStr))
	if !isWithin(s.baseDir, fullPath) {
		return nil, fmt.Errorf("resolved path escapes base directory")
	}
	return os.Open(fullPath) // #nosec G304 — validated by isWithin check above
}

// Delete removes a file by its relative storage path.
func (s *LocalFileStorage) Delete(_ context.Context, pathStr string) error {
	fullPath := filepath.Join(s.baseDir, path.Clean(pathStr))
	if !isWithin(s.baseDir, fullPath) {
		return fmt.Errorf("resolved path escapes base directory")
	}
	return os.Remove(fullPath)
}

// isWithin checks that the target path is inside the base directory.
func isWithin(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !path.IsAbs(rel) && len(rel) > 0
}
