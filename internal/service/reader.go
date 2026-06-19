package service

import (
	"context"
	"fmt"
	"io"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/storage"
)

// CoverResult holds the data returned by ServeCover.
type CoverResult struct {
	Reader io.ReadCloser
}

// DocumentReader handles read operations that absorb domain logic
// beyond a trivial repository call (cover resolution, progress calculation).
type DocumentReader struct {
	docRepo  repository.DocumentRepository
	chRepo   repository.ChapterRepository
	progRepo repository.ReadingProgressRepository
	storage  storage.FileStorage
}

// NewDocumentReader creates a new DocumentReader.
func NewDocumentReader(
	docRepo repository.DocumentRepository,
	chRepo repository.ChapterRepository,
	progRepo repository.ReadingProgressRepository,
	storage storage.FileStorage,
) *DocumentReader {
	return &DocumentReader{
		docRepo:  docRepo,
		chRepo:   chRepo,
		progRepo: progRepo,
		storage:  storage,
	}
}

// ServeCover resolves the document's cover path and opens the cover image.
// This absorbs the two-step dance (get document → get cover) that was previously
// split between handler and service.
func (s *DocumentReader) ServeCover(ctx context.Context, documentID string) (*CoverResult, error) {
	doc, err := s.docRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("serve cover: %w", err)
	}

	if doc.CoverPath == "" {
		return nil, ErrNoCover
	}

	reader, err := s.storage.Get(ctx, doc.CoverPath)
	if err != nil {
		return nil, fmt.Errorf("serve cover: open file: %w", err)
	}

	return &CoverResult{Reader: reader}, nil
}

// SaveProgress saves reading progress for a document.
// It calculates percentage as ((chapterIndex + 1) / totalChapters) * 100.
func (s *DocumentReader) SaveProgress(ctx context.Context, documentID string, chapterIndex int) (*model.ReadingProgress, error) {
	doc, err := s.docRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("save progress: get document: %w", err)
	}

	totalChapters := doc.ChapterCount
	if totalChapters < 1 {
		totalChapters = 1
	}
	percentage := float64(chapterIndex+1) / float64(totalChapters) * 100

	progress, err := s.progRepo.Upsert(documentID, chapterIndex, percentage)
	if err != nil {
		return nil, fmt.Errorf("save progress: %w", err)
	}

	return progress, nil
}
