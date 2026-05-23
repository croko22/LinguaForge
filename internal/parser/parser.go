package parser

import (
	"github.com/croko/language-app/internal/model"
)

// Parser defines the interface for document parsing.
type Parser interface {
	// CanParse checks whether this parser can handle the given file.
	CanParse(filename string) bool

	// Parse extracts structured content from a document file.
	// readerAt is the file content, size is the file size.
	Parse(readerAt ReaderAt, size int64) (*model.ParsedDocument, error)
}

// ReaderAt is the minimal interface needed for parsing.
// Matches io.ReaderAt + io.ReadSeeker.
type ReaderAt interface {
	ReadAt(p []byte, off int64) (int, error)
	Seek(offset int64, whence int) (int64, error)
}
