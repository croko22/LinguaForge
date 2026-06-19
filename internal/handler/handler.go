package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/service"
)

const maxMemoryBuffer = 10 * 1024 * 1024 // 10 MB for in-memory multipart parsing

// DocumentHandler handles HTTP requests for document operations.
// Deep operations (upload, process, delete) go through DocumentIngester.
// Cover and progress operations go through DocumentReader.
// Trivial reads go directly to repositories.
type DocumentHandler struct {
	ingester *service.DocumentIngester
	reader   *service.DocumentReader
	docRepo  repository.DocumentRepository
	chRepo   repository.ChapterRepository
	progRepo repository.ReadingProgressRepository
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(
	ingester *service.DocumentIngester,
	reader *service.DocumentReader,
	docRepo repository.DocumentRepository,
	chRepo repository.ChapterRepository,
	progRepo repository.ReadingProgressRepository,
) *DocumentHandler {
	return &DocumentHandler{
		ingester: ingester,
		reader:   reader,
		docRepo:  docRepo,
		chRepo:   chRepo,
		progRepo: progRepo,
	}
}

// RegisterRoutes registers all document API routes on the given router.
func (h *DocumentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/documents", func(r chi.Router) {
		r.Post("/", h.UploadDocument)                        // POST /api/documents
		r.Get("/", h.ListDocuments)                          // GET /api/documents
		r.Get("/{id}", h.GetDocument)                        // GET /api/documents/{id}
		r.Delete("/{id}", h.DeleteDocument)                  // DELETE /api/documents/{id}
		r.Get("/{id}/chapters", h.ListChapters)              // GET /api/documents/{id}/chapters
		r.Get("/{id}/chapters/{index}", h.GetChapterContent) // GET /api/documents/{id}/chapters/{index}
		r.Get("/{id}/cover", h.ServeCover)                   // GET /api/documents/{id}/cover
		r.Put("/{id}/progress", h.SaveProgress)              // PUT /api/documents/{id}/progress
		r.Get("/{id}/progress", h.GetProgress)               // GET /api/documents/{id}/progress
	})
}

// UploadDocument handles POST /api/documents — uploads a new document.
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxUploadSize)
	if err := r.ParseMultipartForm(maxMemoryBuffer); err != nil { // #nosec G120 — body is bounded by MaxBytesReader above
		respondError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing 'file' field in upload")
		return
	}
	defer file.Close()

	doc, err := h.ingester.UploadDocument(r.Context(), header.Filename, header.Size, file)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFileType):
			respondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrFileTooLarge):
			respondError(w, http.StatusRequestEntityTooLarge, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusAccepted, doc)
}

// ListDocuments handles GET /api/documents — lists all documents.
// Direct repo call: trivial read, no domain logic.
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.docRepo.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if docs == nil {
		docs = make([]*model.DocumentSummary, 0)
	}

	respondJSON(w, http.StatusOK, docs)
}

// GetDocument handles GET /api/documents/{id} — gets a single document.
// Direct repo call: trivial read, no domain logic.
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	doc, err := h.docRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "document not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, doc)
}

// DeleteDocument handles DELETE /api/documents/{id} — deletes a document and its data.
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	if err := h.ingester.DeleteDocument(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "document not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete document")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListChapters handles GET /api/documents/{id}/chapters — lists chapters (summary only).
// Direct repo call: trivial read, no domain logic.
func (h *DocumentHandler) ListChapters(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	chapters, err := h.chRepo.ListByDocumentID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "document not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if chapters == nil {
		chapters = make([]*model.Chapter, 0)
	}

	respondJSON(w, http.StatusOK, chapters)
}

// GetChapterContent handles GET /api/documents/{id}/chapters/{index} — gets full chapter content.
// Direct repo call: trivial read, no domain logic.
func (h *DocumentHandler) GetChapterContent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	indexStr := chi.URLParam(r, "index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chapter index: must be an integer")
		return
	}

	chapter, err := h.chRepo.GetByDocumentAndIndex(r.Context(), id, index)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "chapter not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, chapter)
}

// ServeCover handles GET /api/documents/{id}/cover — serves the cover image.
// Deep operation: resolves document → cover path → opens file. Absorbed into DocumentReader.
func (h *DocumentHandler) ServeCover(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	result, err := h.reader.ServeCover(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNoCover) {
			respondError(w, http.StatusNotFound, "no cover available")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "document not found")
			return
		}
		respondError(w, http.StatusNotFound, "cover not found")
		return
	}
	defer result.Reader.Close()

	buf := make([]byte, 512)
	n, _ := result.Reader.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(buf[:n])
	io.Copy(w, result.Reader)
}

// ── Response helpers ────────────────────────────────────────────────────────────

// respondJSON writes a JSON response with the given status code and data.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("respondJSON: encode error", "error", err)
	}
}

type saveProgressRequest struct {
	ChapterIndex int `json:"chapter_index"`
}

// SaveProgress handles PUT /api/documents/{id}/progress.
// Deep operation: calculates percentage from chapter count. Delegated to DocumentReader.
func (h *DocumentHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	var req saveProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.ChapterIndex < 0 {
		respondError(w, http.StatusBadRequest, "chapter_index must be >= 0")
		return
	}

	progress, err := h.reader.SaveProgress(r.Context(), id, req.ChapterIndex)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "document not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to save progress")
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// GetProgress handles GET /api/documents/{id}/progress.
// Direct repo call: trivial read, no domain logic.
func (h *DocumentHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "document id is required")
		return
	}

	progress, err := h.progRepo.GetByDocumentID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondJSON(w, http.StatusOK, map[string]any{
				"chapter_index": 0,
				"percentage":    0,
			})
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get progress")
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// respondError writes a JSON error response with the given status and message.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
