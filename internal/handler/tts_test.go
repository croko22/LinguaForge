package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/croko/language-app/internal/tts"
)

func TestTTSHandler_MissingWord(t *testing.T) {
	svc := newTestTTSService(t)
	h := NewTTSHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/tts", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTTSHandler_DownstreamFails(t *testing.T) {
	svc := newTestTTSService(t)
	h := NewTTSHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/tts?word=hello&lang=en", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestTTSHandler_CacheHit(t *testing.T) {
	cacheDir := t.TempDir()
	svc, err := tts.NewService(cacheDir, "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	seedCache(t, cacheDir, "hello", "en")

	h := NewTTSHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/tts?word=hello&lang=en", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "audio/mpeg" {
		t.Fatalf("expected audio/mpeg, got %s", ct)
	}

	cache := w.Header().Get("Cache-Control")
	if cache != "public, max-age=86400" {
		t.Fatalf("expected Cache-Control header, got %s", cache)
	}
}

func newTestTTSService(t *testing.T) *tts.Service {
	t.Helper()
	cacheDir := t.TempDir()
	svc, err := tts.NewService(cacheDir, "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return svc
}

func seedCache(t *testing.T, cacheDir, word, lang string) {
	t.Helper()
	voice := tts.VoiceForLanguage(lang)
	name := cacheName(word, lang, voice)
	if err := os.WriteFile(filepath.Join(cacheDir, name), []byte("fake-mp3-data"), 0644); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
}

func cacheName(word, language, voice string) string {
	h := sha256.Sum256([]byte(word + "|" + language + "|" + voice))
	return hex.EncodeToString(h[:]) + ".mp3"
}
