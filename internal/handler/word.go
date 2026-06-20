package handler

import (
	"encoding/json"
	"net/http"

	"github.com/croko/language-app/internal/service"
	"github.com/go-chi/chi/v5"
)

// WordHandler handles CRUD for saved words.
type WordHandler struct {
	svc *service.WordService
}

// NewWordHandler creates a WordHandler.
func NewWordHandler(svc *service.WordService) *WordHandler {
	return &WordHandler{svc: svc}
}

type saveWordRequest struct {
	DocumentID  string `json:"document_id"`
	Word        string `json:"word"`
	Translation string `json:"translation"`
	SourceLang  string `json:"source_lang"`
	TargetLang  string `json:"target_lang"`
}

// SaveWord handles POST /api/words — saves a word with optional auto-translation.
func (h *WordHandler) SaveWord(w http.ResponseWriter, r *http.Request) {
	var req saveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Word == "" {
		respondError(w, http.StatusBadRequest, "word is required")
		return
	}

	saved, err := h.svc.SaveWord(r.Context(), req.DocumentID, req.Word, req.Translation, req.SourceLang, req.TargetLang)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save word")
		return
	}

	respondJSON(w, http.StatusCreated, saved)
}

// ListWords handles GET /api/words — returns all saved words.
func (h *WordHandler) ListWords(w http.ResponseWriter, r *http.Request) {
	words, err := h.svc.ListWords(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list words")
		return
	}

	respondJSON(w, http.StatusOK, words)
}

// DeleteWord handles DELETE /api/words/{id}.
func (h *WordHandler) DeleteWord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "word id is required")
		return
	}

	if err := h.svc.DeleteWord(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete word")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
