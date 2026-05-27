package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/croko/language-app/internal/translator"
)

func TestTranslateHandler_ValidWord(t *testing.T) {
	provider := translator.NewProvider(translator.DefaultSettings())
	h := NewTranslateHandler(provider)

	body := `{"word":"hello","source_lang":"en","target_lang":"es"}`
	req := httptest.NewRequest(http.MethodPost, "/api/translate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Translate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp translator.TranslateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Translation == "" {
		t.Error("expected non-empty translation")
	}
}

func TestTranslateHandler_EmptyWord(t *testing.T) {
	provider := translator.NewProvider(translator.DefaultSettings())
	h := NewTranslateHandler(provider)

	body := `{"word":"","source_lang":"en","target_lang":"es"}`
	req := httptest.NewRequest(http.MethodPost, "/api/translate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Translate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTranslateHandler_InvalidJSON(t *testing.T) {
	provider := translator.NewProvider(translator.DefaultSettings())
	h := NewTranslateHandler(provider)

	req := httptest.NewRequest(http.MethodPost, "/api/translate", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Translate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
