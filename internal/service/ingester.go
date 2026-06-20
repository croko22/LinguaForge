package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/parser"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/storage"
	"github.com/croko/language-app/internal/worker"
)

// DocumentIngester handles document upload, processing, and deletion —
// the deep operations that involve multiple repositories and business rules.
type DocumentIngester struct {
	docRepo    repository.DocumentRepository
	chRepo     repository.ChapterRepository
	progRepo   repository.ReadingProgressRepository
	wordRepo   repository.WordRepository
	reviewRepo repository.ReviewRepository
	storage    storage.FileStorage
	parsers    []parser.Parser
	enqueue    func(worker.Job) error
}

// NewDocumentIngester creates a new DocumentIngester.
func NewDocumentIngester(
	docRepo repository.DocumentRepository,
	chRepo repository.ChapterRepository,
	progRepo repository.ReadingProgressRepository,
	wordRepo repository.WordRepository,
	reviewRepo repository.ReviewRepository,
	storage storage.FileStorage,
	parsers []parser.Parser,
) *DocumentIngester {
	return &DocumentIngester{
		docRepo:    docRepo,
		chRepo:     chRepo,
		progRepo:   progRepo,
		wordRepo:   wordRepo,
		reviewRepo: reviewRepo,
		storage:    storage,
		parsers:    parsers,
	}
}

// SetEnqueueFunc sets the function used to enqueue jobs for async processing.
// If not set, UploadDocument processes books synchronously (backward compat).
func (s *DocumentIngester) SetEnqueueFunc(fn func(worker.Job) error) {
	s.enqueue = fn
}

// UploadDocument uploads, stores, and enqueues a document file for processing.
func (s *DocumentIngester) UploadDocument(ctx context.Context, filename string, fileSize int64, reader io.Reader) (*model.Document, error) {
	var matched parser.Parser
	for _, p := range s.parsers {
		if p.CanParse(filename) {
			matched = p
			break
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFileType, filename)
	}
	if fileSize > MaxUploadSize {
		return nil, fmt.Errorf("%w: got %d bytes", ErrFileTooLarge, fileSize)
	}

	now := time.Now().UTC()
	docID := uuid.New().String()

	doc := &model.Document{
		ID:           docID,
		Title:        filename,
		Filename:     filename,
		FileType:     fileTypeFromFilename(filename),
		FileSize:     fileSize,
		StoragePath:  "",
		Status:       model.StatusPending,
		ErrorMessage: "",
		Language:     "",
		ChapterCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("upload document: create document: %w", err)
	}

	storagePath, err := s.storage.Store(ctx, filename, reader)
	if err != nil {
		_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, "storage failed")
		return nil, fmt.Errorf("upload document: store file: %w", err)
	}
	doc.StoragePath = storagePath

	job := worker.Job{
		DocID:       docID,
		Filename:    filename,
		FileSize:    fileSize,
		StoragePath: storagePath,
	}

	if s.enqueue != nil {
		if err := s.enqueue(job); err != nil {
			_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, "enqueue failed")
			return nil, fmt.Errorf("upload document: enqueue: %w", err)
		}
		return doc, nil
	}

	if err := s.ProcessBook(ctx, job); err != nil {
		return nil, fmt.Errorf("upload document: process: %w", err)
	}

	doc, err = s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("upload document: get after process: %w", err)
	}
	return doc, nil
}

// ProcessBook processes a document in the background.
func (s *DocumentIngester) ProcessBook(ctx context.Context, job worker.Job) error {
	if err := s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusProcessing); err != nil {
		return fmt.Errorf("process book: set processing status: %w", err)
	}

	reader, err := s.storage.Get(ctx, job.StoragePath)
	if err != nil {
		errMsg := fmt.Sprintf("read stored file failed: %s", err.Error())
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, errMsg)
		return fmt.Errorf("process book: read stored file: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		errMsg := fmt.Sprintf("read file content failed: %s", err.Error())
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, errMsg)
		return fmt.Errorf("process book: read file content: %w", err)
	}

	var matched parser.Parser
	for _, p := range s.parsers {
		if p.CanParse(job.Filename) {
			matched = p
			break
		}
	}
	if matched == nil {
		errMsg := fmt.Sprintf("no parser found for file: %s", job.Filename)
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, errMsg)
		return fmt.Errorf("%w: %s", ErrInvalidFileType, job.Filename)
	}

	parsedDoc, err := matched.Parse(bytes.NewReader(data), job.FileSize)
	if err != nil {
		errMsg := fmt.Sprintf("parse failed: %s", err.Error())
		if updateErr := s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, errMsg); updateErr != nil {
			return fmt.Errorf("process book: parse error: %w (status update failed: %v)", err, updateErr)
		}
		return fmt.Errorf("process book: parse: %w", err)
	}
	if len(parsedDoc.Chapters) == 0 {
		errMsg := "parse succeeded but no chapters found"
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, errMsg)
		return fmt.Errorf("process book: %s", errMsg)
	}

	now := time.Now().UTC()
	chapters := make([]*model.Chapter, 0, len(parsedDoc.Chapters))
	for _, pc := range parsedDoc.Chapters {
		chapters = append(chapters, &model.Chapter{
			ID:           uuid.New().String(),
			DocumentID:   job.DocID,
			ChapterIndex: pc.Index,
			ChapterTitle: pc.Title,
			Content:      pc.Content,
			TokenCount:   roughTokenCount(pc.Content),
			CreatedAt:    now,
		})
	}

	if err := s.chRepo.CreateBatch(ctx, chapters); err != nil {
		errMsg := fmt.Sprintf("store chapters failed: %s", err.Error())
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, errMsg)
		return fmt.Errorf("process book: create chapters: %w", err)
	}

	doc, err := s.docRepo.GetByID(ctx, job.DocID)
	if err != nil {
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, "get document failed")
		return fmt.Errorf("process book: get document: %w", err)
	}
	if parsedDoc.Title != "" {
		doc.Title = parsedDoc.Title
	}
	doc.Language = parsedDoc.Language
	doc.ChapterCount = len(chapters)
	doc.UpdatedAt = time.Now().UTC()

	if len(parsedDoc.CoverImageData) > 0 {
		coverFilename := fmt.Sprintf("%s_cover", job.DocID)
		coverPath, err := s.storage.Store(ctx, coverFilename, bytes.NewReader(parsedDoc.CoverImageData))
		if err == nil {
			doc.CoverPath = coverPath
		}
	}

	if err := s.docRepo.UpdateMetadata(ctx, doc); err != nil {
		_ = s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusError, "update metadata failed")
		return fmt.Errorf("process book: update metadata: %w", err)
	}

	if err := s.docRepo.UpdateStatus(ctx, job.DocID, model.StatusReady); err != nil {
		return fmt.Errorf("process book: set ready status: %w", err)
	}

	return nil
}

// DeleteDocument deletes a document and all its associated data (reviews, words,
// reading progress, chapters) in the correct order to respect foreign key constraints,
// then removes stored files.
func (s *DocumentIngester) DeleteDocument(ctx context.Context, id string) error {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	// Cascade deletes respecting FK order:
	// 1. word_reviews (depends on saved_words via word_id FK)
	// 2. saved_words (no FK on document_id, must be explicit)
	// 3. reading_progress (FK cascade exists on documents, explicit for clarity)
	// 4. document_chapters (FK cascade exists on documents, explicit for clarity)
	// 5. document (root entity)
	if err := s.reviewRepo.DeleteByDocumentID(ctx, id); err != nil {
		return fmt.Errorf("delete document: delete word reviews: %w", err)
	}
	if err := s.wordRepo.DeleteByDocumentID(ctx, id); err != nil {
		return fmt.Errorf("delete document: delete saved words: %w", err)
	}
	if err := s.progRepo.DeleteByDocumentID(ctx, id); err != nil {
		return fmt.Errorf("delete document: delete reading progress: %w", err)
	}
	if err := s.chRepo.DeleteByDocumentID(ctx, id); err != nil {
		return fmt.Errorf("delete document: delete chapters: %w", err)
	}
	if err := s.docRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	// Clean up stored files after successful DB deletion
	if doc.StoragePath != "" {
		_ = s.storage.Delete(ctx, doc.StoragePath)
	}
	if doc.CoverPath != "" {
		_ = s.storage.Delete(ctx, doc.CoverPath)
	}

	return nil
}
