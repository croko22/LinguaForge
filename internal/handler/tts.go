package handler

import (
	"net/http"

	"github.com/croko/language-app/internal/tts"
)

// TTSHandler serves MP3 audio for a word via the TTS service.
type TTSHandler struct {
	svc *tts.Service
}

// NewTTSHandler creates a TTSHandler.
func NewTTSHandler(svc *tts.Service) *TTSHandler {
	return &TTSHandler{svc: svc}
}

// ServeHTTP handles GET /api/tts?word=...&lang=... — returns MP3 bytes.
func (h *TTSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		respondError(w, http.StatusBadRequest, "word is required")
		return
	}
	lang := r.URL.Query().Get("lang")

	audio, err := h.svc.Synthesize(r.Context(), word, lang)
	if err != nil {
		respondError(w, http.StatusServiceUnavailable, "tts unavailable")
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio)
}
