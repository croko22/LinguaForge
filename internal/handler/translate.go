package handler

import (
	"encoding/json"
	"net/http"

	"github.com/croko/language-app/internal/translator"
)

type TranslateHandler struct {
	provider *translator.Provider
}

func NewTranslateHandler(provider *translator.Provider) *TranslateHandler {
	return &TranslateHandler{provider: provider}
}

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

	t := h.provider.GetTranslator()
	result, err := t.Translate(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "translation failed")
		return
	}

	respondJSON(w, http.StatusOK, result)
}
