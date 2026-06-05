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
	CoverPath    string    `json:"cover_url,omitempty"`
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
	ErrorMessage string    `json:"error_message,omitempty"`
	Language     string    `json:"language,omitempty"`
	ChapterCount int       `json:"chapter_count"`
	CoverURL     string    `json:"cover_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ParsedDocument is the result of parsing an EPUB file (before DB storage).
type ParsedDocument struct {
	Title          string
	Author         string
	Language       string
	Chapters       []ParsedChapter
	CoverImageData []byte // raw cover image bytes, nil if no cover
}

// ParsedChapter is a single chapter from a parsed EPUB.
type ParsedChapter struct {
	Index   int
	Title   string
	Content string
}

// SavedWord represents a word saved for study.
type SavedWord struct {
	ID          string    `json:"id"`
	DocumentID  string    `json:"document_id"`
	Word        string    `json:"word"`
	Translation string    `json:"translation"`
	SourceLang  string    `json:"source_lang"`
	TargetLang  string    `json:"target_lang"`
	CreatedAt   time.Time `json:"created_at"`
}

// ReviewStatus constants
const (
	ReviewStatusNew       = "new"
	ReviewStatusLearning  = "learning"
	ReviewStatusReview    = "review"
	ReviewStatusSuspended = "suspended"
)

// ReviewCard represents a word's spaced repetition state.
type ReviewCard struct {
	ID             string     `json:"id"`
	WordID         string     `json:"word_id"`
	Status         string     `json:"status"`
	EaseFactor     float64    `json:"ease_factor"`
	IntervalDays   int        `json:"interval_days"`
	Repetitions    int        `json:"repetitions"`
	Lapses         int        `json:"lapses"`
	NextReview     time.Time  `json:"next_review"`
	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Document status constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusError      = "error"
)
