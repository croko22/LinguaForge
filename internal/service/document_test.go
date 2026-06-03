package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/parser"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/storage"
	"github.com/croko/language-app/internal/worker"
)

// ── testParser ──────────────────────────────────────────────────────────────────

type testParser struct {
	canParse bool
	parsed   *model.ParsedDocument
	parseErr error
}

func (p *testParser) CanParse(_ string) bool { return p.canParse }

func (p *testParser) Parse(_ parser.ReaderAt, _ int64) (*model.ParsedDocument, error) {
	return p.parsed, p.parseErr
}

// cannotParseParser always returns false from CanParse.
type cannotParseParser struct{}

func (cannotParseParser) CanParse(_ string) bool { return false }
func (cannotParseParser) Parse(_ parser.ReaderAt, _ int64) (*model.ParsedDocument, error) {
	return nil, fmt.Errorf("should not be called")
}

// ── mockStorage ────────────────────────────────────────────────────────────────

type mockStorage struct {
	storeFunc  func(context.Context, string, io.Reader) (string, error)
	getFunc    func(context.Context, string) (io.ReadCloser, error)
	deleteFunc func(context.Context, string) error
}

func (m *mockStorage) Store(ctx context.Context, filename string, reader io.Reader) (string, error) {
	return m.storeFunc(ctx, filename, reader)
}

func (m *mockStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	return m.getFunc(ctx, path)
}

func (m *mockStorage) Delete(ctx context.Context, path string) error {
	return m.deleteFunc(ctx, path)
}

// ── test helpers ────────────────────────────────────────────────────────────────

type testConfig struct {
	parsers []parser.Parser
	storage storage.FileStorage
	enqueue func(worker.Job) error
}

type testOpt func(*testConfig)

func withEnqueue(fn func(worker.Job) error) testOpt {
	return func(c *testConfig) { c.enqueue = fn }
}

func withParsers(p ...parser.Parser) testOpt {
	return func(c *testConfig) { c.parsers = p }
}

func withStorage(s storage.FileStorage) testOpt {
	return func(c *testConfig) { c.storage = s }
}

func defaultParsers() []parser.Parser {
	return []parser.Parser{
		&testParser{
			canParse: true,
			parsed: &model.ParsedDocument{
				Title:    "Test Book",
				Language: "en",
				Chapters: []model.ParsedChapter{
					{Index: 0, Title: "Chapter 1", Content: "Content of chapter one."},
					{Index: 1, Title: "Chapter 2", Content: "Content of chapter two."},
				},
			},
		},
	}
}

func newTestService(t *testing.T, opts ...testOpt) (*DocumentService, *sql.DB, func()) {
	t.Helper()

	cfg := &testConfig{
		parsers: defaultParsers(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	tempDir, err := os.MkdirTemp("", "service-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
	if err := repository.RunMigrations(db); err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("failed to run migrations: %v", err)
	}

	docRepo := repository.NewDocumentRepository(db)
	chRepo := repository.NewChapterRepository(db)

	var fileStore storage.FileStorage
	if cfg.storage != nil {
		fileStore = cfg.storage
	} else {
		fs, err := storage.NewLocalFileStorage(tempDir)
		if err != nil {
			db.Close()
			os.RemoveAll(tempDir)
			t.Fatalf("failed to create file storage: %v", err)
		}
		fileStore = fs
	}

	svc := NewDocumentService(docRepo, chRepo, fileStore, cfg.parsers)
	if cfg.enqueue != nil {
		svc.SetEnqueueFunc(cfg.enqueue)
	}

	return svc, db, func() {
		db.Close()
		os.RemoveAll(tempDir)
	}
}

func createTestReader() io.Reader {
	return strings.NewReader("dummy epub content")
}

// nopCloser wraps a reader to implement io.ReadCloser.
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

// ── Tests: UploadDocument (async path) ──────────────────────────────────────────

func TestUploadDocumentReturnsPendingStatus(t *testing.T) {
	svc, _, cleanup := newTestService(t, withEnqueue(func(worker.Job) error { return nil }))
	defer cleanup()

	doc, err := svc.UploadDocument(context.Background(), "test.epub", 100, createTestReader())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Status != model.StatusPending {
		t.Fatalf("expected status %q, got %q", model.StatusPending, doc.Status)
	}
}

func TestUploadDocumentCreatesDBRecord(t *testing.T) {
	svc, db, cleanup := newTestService(t, withEnqueue(func(worker.Job) error { return nil }))
	defer cleanup()

	doc, err := svc.UploadDocument(context.Background(), "test.epub", 100, createTestReader())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM documents WHERE id = ?", doc.ID).Scan(&count); err != nil {
		t.Fatalf("failed to query DB: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 document record, got %d", count)
	}
}

func TestUploadDocumentStoresFile(t *testing.T) {
	svc, _, cleanup := newTestService(t, withEnqueue(func(worker.Job) error { return nil }))
	defer cleanup()

	doc, err := svc.UploadDocument(context.Background(), "test.epub", 100, createTestReader())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rc, err := svc.storage.Get(context.Background(), doc.StoragePath)
	if err != nil {
		t.Fatalf("stored file not accessible: %v", err)
	}
	rc.Close()
}

func TestUploadDocumentEnqueueCalledWithCorrectData(t *testing.T) {
	var capturedJob worker.Job
	enqueue := func(job worker.Job) error {
		capturedJob = job
		return nil
	}

	svc, _, cleanup := newTestService(t, withEnqueue(enqueue))
	defer cleanup()

	doc, err := svc.UploadDocument(context.Background(), "test.epub", 100, createTestReader())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedJob.DocID != doc.ID {
		t.Errorf("expected DocID %q, got %q", doc.ID, capturedJob.DocID)
	}
	if capturedJob.Filename != "test.epub" {
		t.Errorf("expected Filename 'test.epub', got %q", capturedJob.Filename)
	}
	if capturedJob.FileSize != 100 {
		t.Errorf("expected FileSize 100, got %d", capturedJob.FileSize)
	}
	if capturedJob.StoragePath != doc.StoragePath {
		t.Errorf("expected StoragePath %q, got %q", doc.StoragePath, capturedJob.StoragePath)
	}
}

// ── Tests: UploadDocument (sync fallback) ──────────────────────────────────────

func TestUploadDocumentFallbackSync(t *testing.T) {
	svc, db, cleanup := newTestService(t)
	defer cleanup()

	doc, err := svc.UploadDocument(context.Background(), "test.epub", 100, createTestReader())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Status != model.StatusReady {
		t.Fatalf("expected status %q, got %q", model.StatusReady, doc.Status)
	}
	if doc.ChapterCount != 2 {
		t.Fatalf("expected chapter_count 2, got %d", doc.ChapterCount)
	}
	if doc.Language != "en" {
		t.Fatalf("expected language 'en', got %q", doc.Language)
	}

	var chapterCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM document_chapters WHERE document_id = ?", doc.ID).Scan(&chapterCount); err != nil {
		t.Fatalf("failed to count chapters: %v", err)
	}
	if chapterCount != 2 {
		t.Fatalf("expected 2 chapters in DB, got %d", chapterCount)
	}
}

// ── Tests: ProcessBook (background processing) ─────────────────────────────────

func TestProcessBookSetsReadyOnSuccess(t *testing.T) {
	svc, db, cleanup := newTestService(t)
	defer cleanup()

	ctx := context.Background()
	docID := uuid.New().String()

	// Insert document with pending status
	now := time.Now().UTC()
	if _, err := db.Exec(`
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		docID, "test.epub", "test.epub", "epub", int64(100), "placeholder", model.StatusPending, now, now,
	); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Store file via service's storage
	storagePath, err := svc.storage.Store(ctx, "test.epub", strings.NewReader("content"))
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}

	// Update storage path in DB
	if _, err := db.Exec("UPDATE documents SET storage_path = ? WHERE id = ?", storagePath, docID); err != nil {
		t.Fatalf("failed to update storage path: %v", err)
	}

	// Process the book
	job := worker.Job{DocID: docID, Filename: "test.epub", FileSize: 100, StoragePath: storagePath}
	if err := svc.ProcessBook(ctx, job); err != nil {
		t.Fatalf("ProcessBook returned error: %v", err)
	}

	// Verify status is ready
	var status string
	if err := db.QueryRow("SELECT status FROM documents WHERE id = ?", docID).Scan(&status); err != nil {
		t.Fatalf("failed to query document: %v", err)
	}
	if status != model.StatusReady {
		t.Fatalf("expected status %q, got %q", model.StatusReady, status)
	}

	// Verify chapters exist
	var chapterCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM document_chapters WHERE document_id = ?", docID).Scan(&chapterCount); err != nil {
		t.Fatalf("failed to count chapters: %v", err)
	}
	if chapterCount != 2 {
		t.Fatalf("expected 2 chapters, got %d", chapterCount)
	}

	// Verify metadata was updated
	var language string
	var title string
	if err := db.QueryRow("SELECT title, COALESCE(language, '') FROM documents WHERE id = ?", docID).Scan(&title, &language); err != nil {
		t.Fatalf("failed to query metadata: %v", err)
	}
	if title != "Test Book" {
		t.Fatalf("expected title 'Test Book', got %q", title)
	}
	if language != "en" {
		t.Fatalf("expected language 'en', got %q", language)
	}
}

func TestProcessBookSetsErrorOnParseFailure(t *testing.T) {
	errParser := &testParser{
		canParse: true,
		parseErr: fmt.Errorf("mock parse error"),
	}
	svc, db, cleanup := newTestService(t, withParsers(errParser))
	defer cleanup()

	ctx := context.Background()
	docID := uuid.New().String()

	now := time.Now().UTC()
	if _, err := db.Exec(`
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		docID, "test.epub", "test.epub", "epub", int64(100), "placeholder", model.StatusPending, now, now,
	); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	storagePath, err := svc.storage.Store(ctx, "test.epub", strings.NewReader("content"))
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}
	if _, err := db.Exec("UPDATE documents SET storage_path = ? WHERE id = ?", storagePath, docID); err != nil {
		t.Fatalf("failed to update storage path: %v", err)
	}

	job := worker.Job{DocID: docID, Filename: "test.epub", FileSize: 100, StoragePath: storagePath}
	if err := svc.ProcessBook(ctx, job); err == nil {
		t.Fatal("expected error, got nil")
	}

	var status, errMsg string
	if err := db.QueryRow("SELECT status, COALESCE(error_message, '') FROM documents WHERE id = ?", docID).Scan(&status, &errMsg); err != nil {
		t.Fatalf("failed to query document: %v", err)
	}
	if status != model.StatusError {
		t.Fatalf("expected status %q, got %q", model.StatusError, status)
	}
	if !strings.Contains(errMsg, "mock parse error") {
		t.Fatalf("expected error message to contain 'mock parse error', got %q", errMsg)
	}
}

func TestProcessBookSetsErrorOnStorageFailure(t *testing.T) {
	fileStore := &mockStorage{
		storeFunc: func(_ context.Context, _ string, _ io.Reader) (string, error) {
			return uuid.New().String() + "_test.epub", nil
		},
		getFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return nil, fmt.Errorf("storage read error")
		},
		deleteFunc: func(_ context.Context, _ string) error { return nil },
	}
	svc, db, cleanup := newTestService(t, withStorage(fileStore))
	defer cleanup()

	ctx := context.Background()
	docID := uuid.New().String()

	now := time.Now().UTC()
	if _, err := db.Exec(`
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		docID, "test.epub", "test.epub", "epub", int64(100), "placeholder", model.StatusPending, now, now,
	); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	storagePath, err := svc.storage.Store(ctx, "test.epub", strings.NewReader("content"))
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}
	if _, err := db.Exec("UPDATE documents SET storage_path = ? WHERE id = ?", storagePath, docID); err != nil {
		t.Fatalf("failed to update storage path: %v", err)
	}

	job := worker.Job{DocID: docID, Filename: "test.epub", FileSize: 100, StoragePath: storagePath}
	if err := svc.ProcessBook(ctx, job); err == nil {
		t.Fatal("expected error, got nil")
	}

	var status, errMsg string
	if err := db.QueryRow("SELECT status, COALESCE(error_message, '') FROM documents WHERE id = ?", docID).Scan(&status, &errMsg); err != nil {
		t.Fatalf("failed to query document: %v", err)
	}
	if status != model.StatusError {
		t.Fatalf("expected status %q, got %q", model.StatusError, status)
	}
	if !strings.Contains(errMsg, "storage read error") {
		t.Fatalf("expected error message to contain 'storage read error', got %q", errMsg)
	}
}

// ── Tests: upload validation ───────────────────────────────────────────────────

func TestUploadDocumentRejectsUnsupportedFileType(t *testing.T) {
	svc, _, cleanup := newTestService(t,
		withParsers(cannotParseParser{}),
		withEnqueue(func(worker.Job) error { return nil }),
	)
	defer cleanup()

	_, err := svc.UploadDocument(context.Background(), "test.xyz", 100, createTestReader())
	if err == nil {
		t.Fatal("expected error for unsupported file type")
	}
	if !strings.Contains(err.Error(), "invalid file type") {
		t.Fatalf("expected 'invalid file type' error, got: %v", err)
	}
}

func TestUploadDocumentRejectsFileTooLarge(t *testing.T) {
	svc, _, cleanup := newTestService(t, withEnqueue(func(worker.Job) error { return nil }))
	defer cleanup()

	_, err := svc.UploadDocument(context.Background(), "test.epub", MaxUploadSize+1, createTestReader())
	if err == nil {
		t.Fatal("expected error for file too large")
	}
	if !strings.Contains(err.Error(), "file too large") {
		t.Fatalf("expected 'file too large' error, got: %v", err)
	}
}

// ── Tests: enqueue error handling ──────────────────────────────────────────────

func TestUploadDocumentEnqueueErrorSetsDBError(t *testing.T) {
	svc, db, cleanup := newTestService(t, withEnqueue(func(worker.Job) error {
		return fmt.Errorf("pool full")
	}))
	defer cleanup()

	_, err := svc.UploadDocument(context.Background(), "test.epub", 100, createTestReader())
	if err == nil {
		t.Fatal("expected error on enqueue failure")
	}
	if !strings.Contains(err.Error(), "enqueue") {
		t.Fatalf("expected error to mention 'enqueue', got: %v", err)
	}

	var status string
	if err := db.QueryRow("SELECT status FROM documents").Scan(&status); err != nil {
		t.Fatalf("failed to query document: %v", err)
	}
	if status != model.StatusError {
		t.Fatalf("expected status %q after enqueue failure, got %q", model.StatusError, status)
	}
}

// ── Tests: ProcessBook edge cases ──────────────────────────────────────────────

func TestProcessBookSetsErrorOnNoParser(t *testing.T) {
	svc, db, cleanup := newTestService(t, withParsers(cannotParseParser{}))
	defer cleanup()

	ctx := context.Background()
	docID := uuid.New().String()

	now := time.Now().UTC()
	if _, err := db.Exec(`
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		docID, "test.epub", "test.epub", "epub", int64(100), "placeholder", model.StatusPending, now, now,
	); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	storagePath, err := svc.storage.Store(ctx, "test.epub", strings.NewReader("content"))
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}
	if _, err := db.Exec("UPDATE documents SET storage_path = ? WHERE id = ?", storagePath, docID); err != nil {
		t.Fatalf("failed to update storage path: %v", err)
	}

	job := worker.Job{DocID: docID, Filename: "test.epub", FileSize: 100, StoragePath: storagePath}
	if err := svc.ProcessBook(ctx, job); err == nil {
		t.Fatal("expected error for missing parser, got nil")
	}

	var status string
	if err := db.QueryRow("SELECT status FROM documents WHERE id = ?", docID).Scan(&status); err != nil {
		t.Fatalf("failed to query document: %v", err)
	}
	if status != model.StatusError {
		t.Fatalf("expected status %q, got %q", model.StatusError, status)
	}
}

// ── Tests: cover image extraction ──────────────────────────────────────────────

func TestProcessBookStoresCover(t *testing.T) {
	coverData := []byte("fake-image-data")
	svc, db, cleanup := newTestService(t, withParsers(&testParser{
		canParse: true,
		parsed: &model.ParsedDocument{
			Title:    "Cover Test",
			Language: "en",
			Chapters: []model.ParsedChapter{
				{Index: 0, Title: "Chapter 1", Content: "Content."},
			},
			CoverImageData: coverData,
		},
	}))
	defer cleanup()

	ctx := context.Background()
	docID := uuid.New().String()

	now := time.Now().UTC()
	if _, err := db.Exec(`
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		docID, "test.epub", "test.epub", "epub", int64(100), "placeholder", model.StatusPending, now, now,
	); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	storagePath, err := svc.storage.Store(ctx, "test.epub", strings.NewReader("content"))
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}

	if _, err := db.Exec("UPDATE documents SET storage_path = ? WHERE id = ?", storagePath, docID); err != nil {
		t.Fatalf("failed to update storage path: %v", err)
	}

	job := worker.Job{DocID: docID, Filename: "test.epub", FileSize: 100, StoragePath: storagePath}
	if err := svc.ProcessBook(ctx, job); err != nil {
		t.Fatalf("ProcessBook returned error: %v", err)
	}

	var coverPath string
	if err := db.QueryRow("SELECT COALESCE(cover_path, '') FROM documents WHERE id = ?", docID).Scan(&coverPath); err != nil {
		t.Fatalf("failed to query cover_path: %v", err)
	}
	if coverPath == "" {
		t.Fatal("expected non-empty cover_path, got empty")
	}

	rc, err := svc.storage.Get(ctx, coverPath)
	if err != nil {
		t.Fatalf("stored cover not accessible: %v", err)
	}
	defer rc.Close()
	storedData, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read stored cover: %v", err)
	}
	if !bytes.Equal(storedData, coverData) {
		t.Fatal("stored cover data does not match original")
	}
}

func TestProcessBookWithoutCover(t *testing.T) {
	svc, db, cleanup := newTestService(t)
	defer cleanup()

	ctx := context.Background()
	docID := uuid.New().String()

	now := time.Now().UTC()
	if _, err := db.Exec(`
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		docID, "test.epub", "test.epub", "epub", int64(100), "placeholder", model.StatusPending, now, now,
	); err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	storagePath, err := svc.storage.Store(ctx, "test.epub", strings.NewReader("content"))
	if err != nil {
		t.Fatalf("failed to store file: %v", err)
	}

	if _, err := db.Exec("UPDATE documents SET storage_path = ? WHERE id = ?", storagePath, docID); err != nil {
		t.Fatalf("failed to update storage path: %v", err)
	}

	job := worker.Job{DocID: docID, Filename: "test.epub", FileSize: 100, StoragePath: storagePath}
	if err := svc.ProcessBook(ctx, job); err != nil {
		t.Fatalf("ProcessBook returned error: %v", err)
	}

	var coverPath string
	if err := db.QueryRow("SELECT COALESCE(cover_path, '') FROM documents WHERE id = ?", docID).Scan(&coverPath); err != nil {
		t.Fatalf("failed to query cover_path: %v", err)
	}
	if coverPath != "" {
		t.Fatalf("expected empty cover_path, got %q", coverPath)
	}
}
