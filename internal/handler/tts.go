package handler

import (
	"net/http"

	"github.com/croko/language-app/internal/tts"
)

type TTSHandler struct {
	svc *tts.Service
}

func NewTTSHandler(svc *tts.Service) *TTSHandler {
	return &TTSHandler{svc: svc}
}

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
