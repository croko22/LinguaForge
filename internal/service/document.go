package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/parser"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/storage"
	"github.com/croko/language-app/internal/worker"
)

// MaxUploadSize is the maximum allowed file size for uploads (50 MB).
const MaxUploadSize int64 = 50 * 1024 * 1024

// Sentinel errors for validation failures.
var (
	ErrInvalidFileType = fmt.Errorf("invalid file type: not supported by any parser")
	ErrFileTooLarge    = fmt.Errorf("file too large: maximum size is 50 MB")
)

// DocumentService provides business logic for document management.
type DocumentService struct {
	docRepo  repository.DocumentRepository
	chRepo   repository.ChapterRepository
	progRepo repository.ReadingProgressRepository
	storage  storage.FileStorage
	parsers  []parser.Parser
	enqueue  func(worker.Job) error
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(
	docRepo repository.DocumentRepository,
	chRepo repository.ChapterRepository,
	progRepo repository.ReadingProgressRepository,
	storage storage.FileStorage,
	parsers []parser.Parser,
) *DocumentService {
	return &DocumentService{
		docRepo:  docRepo,
		chRepo:   chRepo,
		progRepo: progRepo,
		storage:  storage,
		parsers:  parsers,
	}
}

// SetEnqueueFunc sets the function used to enqueue jobs for async processing.
// If not set, UploadDocument processes books synchronously (backward compat).
func (s *DocumentService) SetEnqueueFunc(fn func(worker.Job) error) {
	s.enqueue = fn
}

// UploadDocument uploads, stores, and enqueues a document file for processing.
// When an enqueue function is set, the document is created with "pending" status
// and processing happens asynchronously. Otherwise, processing is synchronous.
func (s *DocumentService) UploadDocument(ctx context.Context, filename string, fileSize int64, reader io.Reader) (*model.Document, error) {
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

	if s.enqueue != nil {
		job := worker.Job{
			DocID:       docID,
			Filename:    filename,
			FileSize:    fileSize,
			StoragePath: storagePath,
		}
		if err := s.enqueue(job); err != nil {
			_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, "enqueue failed")
			return nil, fmt.Errorf("upload document: enqueue: %w", err)
		}
		return doc, nil
	}

	job := worker.Job{
		DocID:       docID,
		Filename:    filename,
		FileSize:    fileSize,
		StoragePath: storagePath,
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
// It reads the stored file, parses it, creates chapter records,
// and updates the document status.
func (s *DocumentService) ProcessBook(ctx context.Context, job worker.Job) error {
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

// ListDocuments returns a summary list of all documents.
func (s *DocumentService) ListDocuments(ctx context.Context) ([]*model.DocumentSummary, error) {
	summaries, err := s.docRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	return summaries, nil
}

// GetDocument returns a document by its ID.
func (s *DocumentService) GetDocument(ctx context.Context, id string) (*model.Document, error) {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return doc, nil
}

// DeleteDocument deletes a document, its chapters, saved words, and review cards.
func (s *DocumentService) DeleteDocument(ctx context.Context, id string) error {
	doc, err := s.docRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if err := s.docRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if doc.StoragePath != "" {
		_ = s.storage.Delete(ctx, doc.StoragePath)
	}
	if doc.CoverPath != "" {
		_ = s.storage.Delete(ctx, doc.CoverPath)
	}

	return nil
}

// GetCover opens a cover image file for reading by its storage path.
func (s *DocumentService) GetCover(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.storage.Get(ctx, path)
}

// GetChapters returns all chapters for a document (without content).
func (s *DocumentService) GetChapters(ctx context.Context, documentID string) ([]*model.Chapter, error) {
	chapters, err := s.chRepo.ListByDocumentID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("get chapters: %w", err)
	}
	return chapters, nil
}

// GetChapterContent returns a specific chapter by document ID and index,
// including the full text content.
func (s *DocumentService) GetChapterContent(ctx context.Context, documentID string, chapterIndex int) (*model.Chapter, error) {
	chapter, err := s.chRepo.GetByDocumentAndIndex(ctx, documentID, chapterIndex)
	if err != nil {
		return nil, fmt.Errorf("get chapter content: %w", err)
	}
	return chapter, nil
}

// SaveProgress saves reading progress for a document.
// It calculates percentage as ((chapterIndex + 1) / totalChapters) * 100.
func (s *DocumentService) SaveProgress(ctx context.Context, documentID string, chapterIndex int) (*model.ReadingProgress, error) {
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

// GetProgress returns reading progress for a document.
func (s *DocumentService) GetProgress(ctx context.Context, documentID string) (*model.ReadingProgress, error) {
	progress, err := s.progRepo.GetByDocumentID(documentID)
	if err != nil {
		return nil, fmt.Errorf("get progress: %w", err)
	}
	return progress, nil
}

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
