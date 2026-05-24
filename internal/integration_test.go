package internal

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "modernc.org/sqlite"

	"github.com/croko/language-app/internal/handler"
	"github.com/croko/language-app/internal/parser"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/service"
	"github.com/croko/language-app/internal/storage"
)

// ── EPUB test helpers ──────────────────────────────────────────────────────────

// createMinimalEPUB builds a valid minimal EPUB file in memory and returns it
// as a bytes.Buffer. The EPUB contains two XHTML chapters with plain text.
func createMinimalEPUB() (*bytes.Buffer, error) {
	var buf bytes.Buffer

	// ── mimetype (must be first, stored without compression) ───────────────
	mimeWriter := zip.NewWriter(&buf)

	// mimetype: stored, not compressed (EPUB requirement)
	mimeHeader := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	}
	mimeFile, err := mimeWriter.CreateHeader(mimeHeader)
	if err != nil {
		return nil, fmt.Errorf("create mimetype entry: %w", err)
	}
	if _, err := mimeFile.Write([]byte("application/epub+zip")); err != nil {
		return nil, fmt.Errorf("write mimetype: %w", err)
	}

	// ── META-INF/container.xml ─────────────────────────────────────────────
	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	if err := writeZipFile(mimeWriter, "META-INF/container.xml", containerXML); err != nil {
		return nil, err
	}

	// ── OEBPS/content.opf ──────────────────────────────────────────────────
	opfXML := `<?xml version="1.0" encoding="UTF-8"?>
<package version="2.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId">
  <metadata>
    <dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Test EPUB</dc:title>
    <dc:creator xmlns:dc="http://purl.org/dc/elements/1.1/" opf:role="aut" xmlns:opf="http://www.idpf.org/2007/opf">Test Author</dc:creator>
    <dc:language xmlns:dc="http://purl.org/dc/elements/1.1/">en</dc:language>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="chapter2" href="chapter2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="chapter1"/>
    <itemref idref="chapter2"/>
  </spine>
</package>`
	if err := writeZipFile(mimeWriter, "OEBPS/content.opf", opfXML); err != nil {
		return nil, err
	}

	// ── OEBPS/toc.ncx ──────────────────────────────────────────────────────
	ncxXML := `<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xmlns="http://www.daisy.org/z3986/2005/ncx/">
  <head>
    <meta name="dtb:uid" content="12345"/>
  </head>
  <docTitle>
    <text>Test EPUB</text>
  </docTitle>
  <navMap>
    <navPoint id="navpoint1" playOrder="1">
      <navLabel><text>Introduction</text></navLabel>
      <content src="chapter1.xhtml"/>
    </navPoint>
    <navPoint id="navpoint2" playOrder="2">
      <navLabel><text>Main Content</text></navLabel>
      <content src="chapter2.xhtml"/>
    </navPoint>
  </navMap>
</ncx>`
	if err := writeZipFile(mimeWriter, "OEBPS/toc.ncx", ncxXML); err != nil {
		return nil, err
	}

	// ── OEBPS/chapter1.xhtml ───────────────────────────────────────────────
	ch1XHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Introduction</title></head>
<body>
  <p>This is the first chapter of the test EPUB document.</p>
  <p>It contains multiple paragraphs of text content for testing purposes.</p>
</body>
</html>`
	if err := writeZipFile(mimeWriter, "OEBPS/chapter1.xhtml", ch1XHTML); err != nil {
		return nil, err
	}

	// ── OEBPS/chapter2.xhtml ───────────────────────────────────────────────
	ch2XHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Main Content</title></head>
<body>
  <p>This is the second chapter.</p>
  <p>It contains different content that should also be extracted.</p>
</body>
</html>`
	if err := writeZipFile(mimeWriter, "OEBPS/chapter2.xhtml", ch2XHTML); err != nil {
		return nil, err
	}

	if err := mimeWriter.Close(); err != nil {
		return nil, fmt.Errorf("close zip writer: %w", err)
	}

	return &buf, nil
}

// writeZipFile adds a text file with the given name and content to a zip writer.
func writeZipFile(w *zip.Writer, name, content string) error {
	f, err := w.Create(name)
	if err != nil {
		return fmt.Errorf("create zip entry %s: %w", name, err)
	}
	if _, err := f.Write([]byte(content)); err != nil {
		return fmt.Errorf("write zip entry %s: %w", name, err)
	}
	return nil
}

// ── Test suite ────────────────────────────────────────────────────────────────

// testDeps holds all shared dependencies for the integration test suite.
type testDeps struct {
	db          *sql.DB
	tempDir     string
	router      *chi.Mux
	fileStorage *storage.LocalFileStorage
}

// setupIntegrationTest creates all dependencies and returns them along with a
// cleanup function. It uses an in-memory SQLite database and a temporary
// directory for file storage.
func setupIntegrationTest(t *testing.T) (*testDeps, func()) {
	t.Helper()

	// ── Temporary directory for uploads ────────────────────────────────────
	tempDir, err := os.MkdirTemp("", "language-app-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// ── In-memory SQLite database ──────────────────────────────────────────
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Enable WAL mode and foreign keys for the in-memory DB (pragmas).
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// ── Run migrations ─────────────────────────────────────────────────────
	if err := repository.RunMigrations(db); err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("failed to run migrations: %v", err)
	}

	// ── Create dependencies ────────────────────────────────────────────────
	docRepo := repository.NewDocumentRepository(db)
	chRepo := repository.NewChapterRepository(db)

	fileStorage, err := storage.NewLocalFileStorage(tempDir)
	if err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create file storage: %v", err)
	}

	epubParser := parser.NewEpubParser()
	docService := service.NewDocumentService(docRepo, chRepo, fileStorage, epubParser)
	docHandler := handler.NewDocumentHandler(docService)

	// ── Chi router ─────────────────────────────────────────────────────────
	r := chi.NewRouter()
	docHandler.RegisterRoutes(r)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	deps := &testDeps{
		db:          db,
		tempDir:     tempDir,
		router:      r,
		fileStorage: fileStorage,
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return deps, cleanup
}

// ── Multipart request helper ───────────────────────────────────────────────────

// createUploadRequest builds an HTTP POST multipart request with a file field.
func createUploadRequest(url string, filename string, fileContent *bytes.Buffer) (*http.Request, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, fileContent); err != nil {
		return nil, fmt.Errorf("copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

// ── Response helpers ───────────────────────────────────────────────────────────

// parseJSON decodes the response body into the given target.
func parseJSON(t *testing.T, body io.Reader, target interface{}) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(target); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
}

// ── Tests ──────────────────────────────────────────────────────────────────────

// TestUploadAndRetrieveEPUB exercises the full upload, list, get,
// and chapter retrieval flow.
func TestUploadAndRetrieveEPUB(t *testing.T) {
	deps, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// ── 1. Generate a test EPUB ────────────────────────────────────────────
	epubContent, err := createMinimalEPUB()
	if err != nil {
		t.Fatalf("failed to create test EPUB: %v", err)
	}

	// ── 2. POST /api/documents ─────────────────────────────────────────────
	uploadReq, err := createUploadRequest("/api/documents", "test-epub.epub", epubContent)
	if err != nil {
		t.Fatalf("failed to create upload request: %v", err)
	}

	uploadRec := httptest.NewRecorder()
	deps.router.ServeHTTP(uploadRec, uploadReq)

	// ── 3. Assert upload response ──────────────────────────────────────────
	if uploadRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d; body: %s", uploadRec.Code, uploadRec.Body.String())
	}

	var uploadResp map[string]interface{}
	parseJSON(t, uploadRec.Body, &uploadResp)

	if id, ok := uploadResp["id"].(string); !ok || id == "" {
		t.Fatal("expected non-empty 'id' in upload response")
	}
	if title, ok := uploadResp["title"].(string); !ok || title == "" {
		t.Fatal("expected non-empty 'title' in upload response")
	}
	if status, ok := uploadResp["status"].(string); !ok || status != "ready" {
		t.Fatalf("expected status 'ready', got %v", uploadResp["status"])
	}
	if cc, ok := uploadResp["chapter_count"].(float64); !ok || cc != 2 {
		t.Fatalf("expected chapter_count 2, got %v", uploadResp["chapter_count"])
	}

	docID, _ := uploadResp["id"].(string)
	docTitle, _ := uploadResp["title"].(string)

	// ── 4. GET /api/documents — list ───────────────────────────────────────
	listReq := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	listRec := httptest.NewRecorder()
	deps.router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", listRec.Code, listRec.Body.String())
	}

	var listResp []map[string]interface{}
	parseJSON(t, listRec.Body, &listResp)
	if len(listResp) != 1 {
		t.Fatalf("expected 1 document in list, got %d", len(listResp))
	}

	// ── 5. GET /api/documents/{id} — get single document ───────────────────
	getReq := httptest.NewRequest(http.MethodGet, "/api/documents/"+docID, nil)
	getRec := httptest.NewRecorder()
	deps.router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get document, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var getResp map[string]interface{}
	parseJSON(t, getRec.Body, &getResp)

	if id, ok := getResp["id"].(string); !ok || id != docID {
		t.Fatalf("expected id %q in get response", docID)
	}
	if title, ok := getResp["title"].(string); !ok || title != docTitle {
		t.Fatalf("expected title %q in get response", docTitle)
	}
	if status, ok := getResp["status"].(string); !ok || status != "ready" {
		t.Fatalf("expected status 'ready', got %v", getResp["status"])
	}
	if lang, ok := getResp["language"].(string); !ok || lang != "en" {
		t.Fatalf("expected language 'en', got %v", getResp["language"])
	}

	// ── 6. GET /api/documents/{id}/chapters — list chapters ────────────────
	chaptersReq := httptest.NewRequest(http.MethodGet, "/api/documents/"+docID+"/chapters", nil)
	chaptersRec := httptest.NewRecorder()
	deps.router.ServeHTTP(chaptersRec, chaptersReq)

	if chaptersRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list chapters, got %d: %s", chaptersRec.Code, chaptersRec.Body.String())
	}

	var chaptersResp []map[string]interface{}
	parseJSON(t, chaptersRec.Body, &chaptersResp)
	if len(chaptersResp) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chaptersResp))
	}

	// ── 7. GET /api/documents/{id}/chapters/0 — get first chapter content ──
	ch0Req := httptest.NewRequest(http.MethodGet, "/api/documents/"+docID+"/chapters/0", nil)
	ch0Rec := httptest.NewRecorder()
	deps.router.ServeHTTP(ch0Rec, ch0Req)

	if ch0Rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get chapter 0, got %d: %s", ch0Rec.Code, ch0Rec.Body.String())
	}

	var ch0Resp map[string]interface{}
	parseJSON(t, ch0Rec.Body, &ch0Resp)

	content, ok := ch0Resp["content"].(string)
	if !ok {
		t.Fatal("expected 'content' field in chapter response")
	}
	if content == "" {
		t.Fatal("expected non-empty chapter content")
	}
	if strings.ContainsAny(content, "<>") {
		t.Fatal("expected chapter content to contain text without HTML tags")
	}
	if !strings.Contains(content, "first chapter") {
		t.Fatalf("expected chapter content to contain 'first chapter', got: %s", content)
	}

	// ── 8. GET /api/documents/{id}/chapters/99 — out of bounds → 404 ───────
	ch99Req := httptest.NewRequest(http.MethodGet, "/api/documents/"+docID+"/chapters/99", nil)
	ch99Rec := httptest.NewRecorder()
	deps.router.ServeHTTP(ch99Rec, ch99Req)

	if ch99Rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for out-of-bounds chapter, got %d: %s", ch99Rec.Code, ch99Rec.Body.String())
	}
}

// TestUploadInvalidFile verifies that uploading a non-EPUB file
// returns a 400 error.
func TestUploadInvalidFile(t *testing.T) {
	deps, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create a plain text file (not an EPUB).
	fileContent := bytes.NewBufferString("This is not an epub file, it's just a text file.")
	uploadReq, err := createUploadRequest("/api/documents", "test.txt", fileContent)
	if err != nil {
		t.Fatalf("failed to create upload request: %v", err)
	}

	uploadRec := httptest.NewRecorder()
	deps.router.ServeHTTP(uploadRec, uploadReq)

	if uploadRec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid file, got %d: %s", uploadRec.Code, uploadRec.Body.String())
	}

	var errResp map[string]string
	parseJSON(t, uploadRec.Body, &errResp)

	if msg, ok := errResp["error"]; !ok || msg == "" {
		t.Fatal("expected non-empty 'error' field in response")
	}
}

// TestHealthEndpoint verifies the health check endpoint.
func TestHealthEndpoint(t *testing.T) {
	deps, cleanup := setupIntegrationTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	deps.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var healthResp map[string]string
	parseJSON(t, rec.Body, &healthResp)
	if status, ok := healthResp["status"]; !ok || status != "ok" {
		t.Fatalf("expected status 'ok', got %v", healthResp["status"])
	}
}
