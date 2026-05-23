package storage

import (
	"context"
	"io"
)

// FileStorage abstracts file persistence.
type FileStorage interface {
	// Store saves a file and returns its storage path.
	Store(ctx context.Context, filename string, reader io.Reader) (path string, err error)

	// Get opens a file for reading by its storage path.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file by its storage path.
	Delete(ctx context.Context, path string) error
}
