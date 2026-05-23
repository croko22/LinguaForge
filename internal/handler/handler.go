package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/service"
)

const maxMemoryBuffer = 10 * 1024 * 1024 // 10 MB for in-memory multipart parsing

// DocumentHandler handles HTTP requests for document operations.
type DocumentHandler struct {
	svc *service.DocumentService
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

// RegisterRoutes registers all document API routes on the given router.
func (h *DocumentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/documents", func(r chi.Router) {
		r.Post("/", h.UploadDocument)                                  // POST /api/documents
		r.Get("/", h.ListDocuments)                                    // GET /api/documents
		r.Get("/{id}", h.GetDocument)                                  // GET /api/documents/{id}
		r.Get("/{id}/chapters", h.ListChapters)                        // GET /api/documents/{id}/chapters
		r.Get("/{id}/chapters/{index}", h.GetChapterContent)           // GET /api/documents/{id}/chapters/{index}
	})
}

// UploadDocument handles POST /api/documents — uploads a new document.
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxMemoryBuffer); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing 'file' field in upload")
		return
	}
	defer file.Close()

	doc, err := h.svc.UploadDocument(r.Context(), header.Filename, header.Size, file)
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

	respondJSON(w, http.StatusCreated, doc)
}

// ListDocuments handles GET /api/documents — lists all documents.
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.svc.ListDocuments(r.Context())
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
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	doc, err := h.svc.GetDocument(r.Context(), id)
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

// ListChapters handles GET /api/documents/{id}/chapters — lists chapters (summary only).
func (h *DocumentHandler) ListChapters(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	chapters, err := h.svc.GetChapters(r.Context(), id)
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
func (h *DocumentHandler) GetChapterContent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	indexStr := chi.URLParam(r, "index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chapter index: must be an integer")
		return
	}

	chapter, err := h.svc.GetChapterContent(r.Context(), id, index)
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

// ── Response helpers ────────────────────────────────────────────────────────────

// respondJSON writes a JSON response with the given status code and data.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes a JSON error response with the given status and message.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
