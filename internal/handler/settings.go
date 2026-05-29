package handler

import (
	"encoding/json"
	"net/http"

	"github.com/croko/language-app/internal/translator"
)

// SettingsHandler handles GET/PUT /api/settings.
type SettingsHandler struct {
	provider *translator.Provider
}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler(provider *translator.Provider) *SettingsHandler {
	return &SettingsHandler{provider: provider}
}

// GetSettings returns the current translation settings.
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, h.provider.GetSettings())
}

// UpdateSettings updates the translation settings and recreates the active provider.
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var s translator.Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if !providerExists(s.ActiveProvider, s.Providers) {
		respondError(w, http.StatusBadRequest, "unknown provider: "+s.ActiveProvider)
		return
	}

	h.provider.ApplySettings(&s)

	respondJSON(w, http.StatusOK, h.provider.GetSettings())
}

// providerExists checks that activeProvider is present in the providers list.
func providerExists(activeProvider string, providers []translator.ProviderConfig) bool {
	for _, p := range providers {
		if p.Name == activeProvider {
			return true
		}
	}
	return false
}
