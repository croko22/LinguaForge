package handler

import (
	"encoding/json"
	"net/http"

	"github.com/croko/language-app/internal/translator"
)

// TranslateHandler handles word translation requests.
type TranslateHandler struct {
	translator translator.Translator
}

// NewTranslateHandler creates a TranslateHandler.
func NewTranslateHandler(t translator.Translator) *TranslateHandler {
	return &TranslateHandler{translator: t}
}

// Translate handles POST /api/translate — accepts a word and returns its translation.
func (h *TranslateHandler) Translate(w http.ResponseWriter, r *http.Request) {
	var req translator.TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Word == "" {
		respondError(w, http.StatusBadRequest, "word cannot be empty")
		return
	}

	result, err := h.translator.Translate(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "translation failed")
		return
	}

	respondJSON(w, http.StatusOK, result)
}
