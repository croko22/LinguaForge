package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/service"
	"github.com/go-chi/chi/v5"
)

// withChiContext embeds chi URL params in the request context for testing.
func withChiContext(ctx context.Context, rctx *chi.Context) context.Context {
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

func setupReviewHandler(t *testing.T) (*ReviewHandler, *repository.ReviewRepoMock, *repository.WordRepoMock) {
	t.Helper()
	reviewMock := repository.NewReviewRepoMock()
	wordMock := repository.NewWordRepoMock()
	svc := service.NewReviewService(reviewMock, wordMock)
	h := NewReviewHandler(svc)
	return h, reviewMock, wordMock
}

func TestReviewHandler_GetDueWords_Empty(t *testing.T) {
	h, _, _ := setupReviewHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/words/due", nil)
	w := httptest.NewRecorder()

	h.GetDueWords(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp dueWordsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Count != 0 {
		t.Fatalf("expected 0, got %d", resp.Count)
	}
	if len(resp.Words) != 0 {
		t.Fatalf("expected empty words, got %d", len(resp.Words))
	}
}

func TestReviewHandler_GetDueWords_WithData(t *testing.T) {
	h, reviewMock, wordMock := setupReviewHandler(t)
	now := time.Now().UTC()

	reviewMock.Seed([]*model.ReviewCard{
		{WordID: "w1", Status: "review", NextReview: now, EaseFactor: 2.5, IntervalDays: 1, Repetitions: 1},
		{WordID: "w2", Status: "new", NextReview: now, EaseFactor: 2.5, IntervalDays: 0, Repetitions: 0},
	})
	wordMock.Seed([]*model.SavedWord{
		{ID: "w1", Word: "hello", Translation: "hola", SourceLang: "en", TargetLang: "es"},
		{ID: "w2", Word: "world", Translation: "mundo", SourceLang: "en", TargetLang: "es"},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/words/due", nil)
	w := httptest.NewRecorder()

	h.GetDueWords(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp dueWordsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Count != 2 {
		t.Fatalf("expected 2, got %d", resp.Count)
	}
	if len(resp.Words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(resp.Words))
	}
	if resp.Words[0].Word == "" {
		t.Fatal("expected non-empty word")
	}
}

func TestReviewHandler_SubmitReview_Success(t *testing.T) {
	h, reviewMock, _ := setupReviewHandler(t)
	now := time.Now().UTC()

	reviewMock.Seed([]*model.ReviewCard{
		{WordID: "w1", Status: "new", NextReview: now, EaseFactor: 2.5, IntervalDays: 0, Repetitions: 0, Lapses: 0},
	})

	body := `{"quality":3}`
	req := httptest.NewRequest(http.MethodPost, "/api/words/w1/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w1")
	req = req.WithContext(withChiContext(req.Context(), rctx))

	w := httptest.NewRecorder()
	h.SubmitReview(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var card model.ReviewCard
	if err := json.NewDecoder(w.Body).Decode(&card); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if card.Status != "learning" {
		t.Fatalf("expected learning, got %s", card.Status)
	}
}

func TestReviewHandler_SubmitReview_MissingID(t *testing.T) {
	h, _, _ := setupReviewHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/words//review", strings.NewReader(`{"quality":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SubmitReview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReviewHandler_SubmitReview_InvalidJSON(t *testing.T) {
	h, _, _ := setupReviewHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/words/w1/review", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w1")
	req = req.WithContext(withChiContext(req.Context(), rctx))

	w := httptest.NewRecorder()
	h.SubmitReview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReviewHandler_SubmitReview_InvalidQuality(t *testing.T) {
	h, _, _ := setupReviewHandler(t)
	body := `{"quality":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/words/w1/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "w1")
	req = req.WithContext(withChiContext(req.Context(), rctx))

	w := httptest.NewRecorder()
	h.SubmitReview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReviewHandler_SubmitReview_WordNotFound(t *testing.T) {
	h, _, _ := setupReviewHandler(t)
	body := `{"quality":3}`
	req := httptest.NewRequest(http.MethodPost, "/api/words/nonexistent/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(withChiContext(req.Context(), rctx))

	w := httptest.NewRecorder()
	h.SubmitReview(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReviewHandler_CountDue(t *testing.T) {
	h, reviewMock, _ := setupReviewHandler(t)
	now := time.Now().UTC()

	reviewMock.Seed([]*model.ReviewCard{
		{WordID: "w1", Status: "new", NextReview: now, EaseFactor: 2.5},
		{WordID: "w2", Status: "review", NextReview: now, EaseFactor: 2.5},
		{WordID: "w3", Status: "learning", NextReview: now, EaseFactor: 2.5},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/words/due/count", nil)
	w := httptest.NewRecorder()

	h.CountDue(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["count"] != 3 {
		t.Fatalf("expected 3, got %d", resp["count"])
	}
}
