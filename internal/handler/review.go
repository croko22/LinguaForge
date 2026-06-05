package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/croko/language-app/internal/service"
	"github.com/croko/language-app/internal/srs"
	"github.com/go-chi/chi/v5"
)

// ReviewHandler handles HTTP requests for spaced repetition review operations.
type ReviewHandler struct {
	svc *service.ReviewService
}

// NewReviewHandler creates a new ReviewHandler.
func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// RegisterRoutes registers all review API routes on the given router.
func (h *ReviewHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/words/due", h.GetDueWords)
	r.Get("/api/words/due/count", h.CountDue)
	r.Post("/api/words/{id}/review", h.SubmitReview)
}

type dueWordsResponse struct {
	Words []*service.DueWordResponse `json:"words"`
	Count int                        `json:"count"`
}

// GetDueWords handles GET /api/words/due — returns due review cards with word data.
func (h *ReviewHandler) GetDueWords(w http.ResponseWriter, r *http.Request) {
	words, err := h.svc.GetDue(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get due words")
		return
	}

	respondJSON(w, http.StatusOK, dueWordsResponse{
		Words: words,
		Count: len(words),
	})
}

type submitReviewRequest struct {
	Quality int `json:"quality"`
}

// SubmitReview handles POST /api/words/{id}/review — processes a review submission.
func (h *ReviewHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	wordID := chi.URLParam(r, "id")
	if wordID == "" {
		respondError(w, http.StatusBadRequest, "word id is required")
		return
	}

	var req submitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Quality < 1 || req.Quality > 4 {
		respondError(w, http.StatusBadRequest, "quality must be between 1 and 4")
		return
	}

	card, err := h.svc.SubmitReview(r.Context(), wordID, srs.Quality(req.Quality))
	if err != nil {
		if errors.Is(err, service.ErrInvalidQuality) {
			respondError(w, http.StatusBadRequest, "quality must be between 1 and 4")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "word not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to submit review")
		return
	}

	respondJSON(w, http.StatusOK, card)
}

// CountDue handles GET /api/words/due/count — returns the count of due review cards.
func (h *ReviewHandler) CountDue(w http.ResponseWriter, r *http.Request) {
	count, err := h.svc.CountDue(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to count due words")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"count": count})
}
