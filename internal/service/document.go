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
	docRepo repository.DocumentRepository
	chRepo  repository.ChapterRepository
	storage storage.FileStorage
	parser  parser.Parser
}

// NewDocumentService creates a new DocumentService.
func NewDocumentService(
	docRepo repository.DocumentRepository,
	chRepo repository.ChapterRepository,
	storage storage.FileStorage,
	parser parser.Parser,
) *DocumentService {
	return &DocumentService{
		docRepo: docRepo,
		chRepo:  chRepo,
		storage: storage,
		parser:  parser,
	}
}

// UploadDocument uploads, stores, and parses a document file.
// It returns the fully populated Document on success, or an error on failure.
func (s *DocumentService) UploadDocument(ctx context.Context, filename string, fileSize int64, reader io.Reader) (*model.Document, error) {
	// ── Validation ───────────────────────────────────────────────────────────
	if !s.parser.CanParse(filename) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFileType, filename)
	}
	if fileSize > MaxUploadSize {
		return nil, fmt.Errorf("%w: got %d bytes", ErrFileTooLarge, fileSize)
	}

	now := time.Now().UTC()
	docID := uuid.New().String()

	// ── Create document record with "processing" status ──────────────────────
	doc := &model.Document{
		ID:           docID,
		Title:        filename,
		Filename:     filename,
		FileType:     fileTypeFromFilename(filename),
		FileSize:     fileSize,
		StoragePath:  "",
		Status:       model.StatusProcessing,
		ErrorMessage: "",
		Language:     "",
		ChapterCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("upload document: create document: %w", err)
	}

	// ── Store file via storage ───────────────────────────────────────────────
	storagePath, err := s.storage.Store(ctx, filename, reader)
	if err != nil {
		_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, "storage failed")
		return nil, fmt.Errorf("upload document: store file: %w", err)
	}
	doc.StoragePath = storagePath

	// ── Read stored file for parsing ─────────────────────────────────────────
	readerCloser, err := s.storage.Get(ctx, storagePath)
	if err != nil {
		_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, "read stored file failed")
		return nil, fmt.Errorf("upload document: read stored file: %w", err)
	}
	defer readerCloser.Close()

	data, err := io.ReadAll(readerCloser)
	if err != nil {
		_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, "read file content failed")
		return nil, fmt.Errorf("upload document: read file content: %w", err)
	}

	// ── Parse the document ───────────────────────────────────────────────────
	parsedDoc, err := s.parser.Parse(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		errMsg := fmt.Sprintf("parse failed: %s", err.Error())
		if updateErr := s.docRepo.UpdateStatus(ctx, docID, model.StatusError, errMsg); updateErr != nil {
			return nil, fmt.Errorf("upload document: parse error: %w (status update failed: %v)", err, updateErr)
		}
		return nil, fmt.Errorf("upload document: parse: %w", err)
	}

	// ── Build chapter records ────────────────────────────────────────────────
	chapters := make([]*model.Chapter, 0, len(parsedDoc.Chapters))
	for _, pc := range parsedDoc.Chapters {
		chapters = append(chapters, &model.Chapter{
			ID:           uuid.New().String(),
			DocumentID:   docID,
			ChapterIndex: pc.Index,
			ChapterTitle: pc.Title,
			Content:      pc.Content,
			TokenCount:   roughTokenCount(pc.Content),
			CreatedAt:    now,
		})
	}

	// ── Batch-create chapters ────────────────────────────────────────────────
	if err := s.chRepo.CreateBatch(ctx, chapters); err != nil {
		errMsg := fmt.Sprintf("store chapters failed: %s", err.Error())
		_ = s.docRepo.UpdateStatus(ctx, docID, model.StatusError, errMsg)
		return nil, fmt.Errorf("upload document: create chapters: %w", err)
	}

	// ── Update document status to "ready" ────────────────────────────────────
	if err := s.docRepo.UpdateStatus(ctx, docID, model.StatusReady); err != nil {
		return nil, fmt.Errorf("upload document: update status to ready: %w", err)
	}

	// Populate the returned document with parsed metadata.
	doc.Title = parsedDoc.Title
	doc.Language = parsedDoc.Language
	doc.ChapterCount = len(chapters)
	doc.Status = model.StatusReady
	doc.UpdatedAt = time.Now().UTC()

	return doc, nil
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

// ── Helpers ───────────────────────────────────────────────────────────────────

// fileTypeFromFilename extracts the file extension without the leading dot.
func fileTypeFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimPrefix(strings.ToLower(ext), ".")
}

// roughTokenCount estimates the number of tokens in a text string.
// Uses whitespace-delimited word count as a reasonable approximation.
func roughTokenCount(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Fields(content))
}
