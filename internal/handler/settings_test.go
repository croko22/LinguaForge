package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/croko/language-app/internal/translator"
)

func TestSettingsHandler_GetSettings(t *testing.T) {
	settings := translator.DefaultSettings()
	provider := translator.NewProvider(settings)
	h := NewSettingsHandler(provider)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	h.GetSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp translator.Settings
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ActiveProvider != "mock" {
		t.Fatalf("expected active_provider 'mock', got '%s'", resp.ActiveProvider)
	}
	if len(resp.Providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(resp.Providers))
	}
}

func TestSettingsHandler_UpdateSettings(t *testing.T) {
	settings := translator.DefaultSettings()
	provider := translator.NewProvider(settings)
	h := NewSettingsHandler(provider)

	body := `{"active_provider":"libre","providers":[{"name":"mock"},{"name":"libre","endpoint":"http://test:5000"}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify settings were stored
	updated := provider.GetSettings()
	if updated.ActiveProvider != "libre" {
		t.Fatalf("expected active_provider 'libre', got '%s'", updated.ActiveProvider)
	}

	// Verify it also switched the translator (should now be libre, not mock)
	trans := provider.GetTranslator()
	if _, ok := trans.(*translator.LibreClient); !ok {
		t.Fatal("expected translator to be a LibreClient after updating settings")
	}
}

func TestSettingsHandler_UpdateSettings_InvalidProvider(t *testing.T) {
	settings := translator.DefaultSettings()
	provider := translator.NewProvider(settings)
	h := NewSettingsHandler(provider)

	body := `{"active_provider":"nonexistent","providers":[{"name":"mock"},{"name":"libre"}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown provider, got %d: %s", w.Code, w.Body.String())
	}

	// Verify settings were NOT changed
	updated := provider.GetSettings()
	if updated.ActiveProvider != "mock" {
		t.Fatalf("expected active_provider to remain 'mock', got '%s'", updated.ActiveProvider)
	}
}

func TestSettingsHandler_UpdateSettings_InvalidJSON(t *testing.T) {
	settings := translator.DefaultSettings()
	provider := translator.NewProvider(settings)
	h := NewSettingsHandler(provider)

	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsHandler_ProviderSwitchesTranslator(t *testing.T) {
	settings := translator.DefaultSettings()
	provider := translator.NewProvider(settings)
	_ = NewSettingsHandler(provider)

	// Before: mock
	trans := provider.GetTranslator()
	if _, ok := trans.(*translator.MockTranslator); !ok {
		t.Fatal("expected mock translator by default")
	}

	// After: libre
	provider.ApplySettings(&translator.Settings{
		ActiveProvider: "libre",
		Providers: []translator.ProviderConfig{
			{Name: "mock"},
			{Name: "libre", Endpoint: "http://test:5000"},
		},
	})

	trans = provider.GetTranslator()
	if _, ok := trans.(*translator.LibreClient); !ok {
		t.Fatal("expected libre translator after switching provider")
	}
}
