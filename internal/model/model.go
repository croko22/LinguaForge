package model

import "time"

// Document represents an uploaded EPUB document.
type Document struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Filename     string    `json:"filename"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	StoragePath  string    `json:"-"`
	Status       string    `json:"status"` // pending, processing, ready, error
	ErrorMessage string    `json:"error_message,omitempty"`
	Language     string    `json:"language,omitempty"`
	ChapterCount int       `json:"chapter_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Chapter represents a single chapter extracted from a document.
type Chapter struct {
	ID           string    `json:"id"`
	DocumentID   string    `json:"document_id"`
	ChapterIndex int       `json:"chapter_index"`
	ChapterTitle string    `json:"chapter_title"`
	Content      string    `json:"content"`
	TokenCount   int       `json:"token_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// DocumentSummary is the lightweight view returned in document lists.
type DocumentSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	Status       string    `json:"status"`
	Language     string    `json:"language,omitempty"`
	ChapterCount int       `json:"chapter_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// ParsedDocument is the result of parsing an EPUB file (before DB storage).
type ParsedDocument struct {
	Title    string
	Author   string
	Language string
	Chapters []ParsedChapter
}

// ParsedChapter is a single chapter from a parsed EPUB.
type ParsedChapter struct {
	Index   int
	Title   string
	Content string
}

// Document status constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusError      = "error"
)
