package handler

import (
	"net/http"

	"github.com/croko/language-app/internal/repository"
)

type StatsHandler struct {
	docRepo    repository.DocumentRepository
	chRepo     repository.ChapterRepository
	wordRepo   repository.WordRepository
	reviewRepo repository.ReviewRepository
}

func NewStatsHandler(
	docRepo repository.DocumentRepository,
	chRepo repository.ChapterRepository,
	wordRepo repository.WordRepository,
	reviewRepo repository.ReviewRepository,
) *StatsHandler {
	return &StatsHandler{
		docRepo:    docRepo,
		chRepo:     chRepo,
		wordRepo:   wordRepo,
		reviewRepo: reviewRepo,
	}
}

type statsResponse struct {
	TotalDocuments int                         `json:"total_documents"`
	TotalWords     int                         `json:"total_words"`
	TotalChapters  int                         `json:"total_chapters"`
	LanguageCounts []repository.LanguageCount  `json:"language_counts"`
	ReviewActivity []repository.ReviewActivity `json:"review_activity"`
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	docCount, err := h.docRepo.CountAll(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get document count")
		return
	}

	chCount, err := h.chRepo.CountAll(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get chapter count")
		return
	}

	wordCount, err := h.wordRepo.CountAll(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get word count")
		return
	}

	langCounts, err := h.docRepo.CountByLanguage(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get language counts")
		return
	}

	activity, err := h.reviewRepo.GetReviewActivity(ctx, 365)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get review activity")
		return
	}

	respondJSON(w, http.StatusOK, statsResponse{
		TotalDocuments: docCount,
		TotalWords:     wordCount,
		TotalChapters:  chCount,
		LanguageCounts: langCounts,
		ReviewActivity: activity,
	})
}
