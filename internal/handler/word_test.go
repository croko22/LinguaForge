package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/service"
)

func setupWordHandler(t *testing.T) (*WordHandler, *repository.WordRepoMock) {
	t.Helper()
	mock := repository.NewWordRepoMock()
	svc := service.NewWordService(mock)
	h := NewWordHandler(svc)
	return h, mock
}

func TestWordHandler_SaveWord(t *testing.T) {
	h, _ := setupWordHandler(t)

	body := `{"document_id":"doc-1","word":"hello","translation":"hola","source_lang":"en","target_lang":"es"}`
	req := httptest.NewRequest(http.MethodPost, "/api/words", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SaveWord(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp repository.SavedWord
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Word != "hello" {
		t.Fatalf("expected 'hello', got '%s'", resp.Word)
	}
	if resp.Translation != "hola" {
		t.Fatalf("expected 'hola', got '%s'", resp.Translation)
	}
}

func TestWordHandler_ListWords(t *testing.T) {
	h, mock := setupWordHandler(t)
	mock.Seed([]*repository.SavedWord{
		{ID: "1", Word: "hello", Translation: "hola"},
		{ID: "2", Word: "world", Translation: "mundo"},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/words", nil)
	w := httptest.NewRecorder()

	h.ListWords(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var words []*repository.SavedWord
	json.NewDecoder(w.Body).Decode(&words)
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
}

func TestWordHandler_DeleteWord(t *testing.T) {
	h, _ := setupWordHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/words/w1", nil)
	w := httptest.NewRecorder()

	h.DeleteWord(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
