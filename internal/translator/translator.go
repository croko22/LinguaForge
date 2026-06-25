package translator

import (
	"context"
	"sync"
)

// TranslateRequest is the payload for a translation request.
type TranslateRequest struct {
	Word       string `json:"word"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// TranslateResponse holds the result of a translation.
type TranslateResponse struct {
	Translation string `json:"translation"`
}

// Translator translates a single word from one language to another.
type Translator interface {
	Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error)
}

// NewFromSettings creates a Translator based on the active provider in settings.
func NewFromSettings(settings *Settings) Translator {
	for _, p := range settings.Providers {
		if p.Name == settings.ActiveProvider {
			switch p.Name {
			case "libre":
				return NewLibreTranslate(p.Endpoint, p.APIKey)
			case "deepl":
				return NewDeepLTranslate(p.Endpoint, p.APIKey)
			case "mock":
				return NewMockTranslator()
			}
		}
	}
	return NewMockTranslator()
}

// Provider manages the current translation settings and active translator.
// It is safe for concurrent use.
type Provider struct {
	mu       sync.RWMutex
	settings *Settings
	current  Translator
}

// NewProvider creates a Provider from the given settings.
func NewProvider(settings *Settings) *Provider {
	return &Provider{
		settings: settings,
		current:  NewFromSettings(settings),
	}
}

// Translate delegates to the active translator.
func (p *Provider) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
	return p.GetTranslator().Translate(ctx, req)
}

// GetTranslator returns the currently active Translator.
func (p *Provider) GetTranslator() Translator {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

// ApplySettings updates the settings and recreates the active translator.
func (p *Provider) ApplySettings(s *Settings) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.settings = s
	p.current = NewFromSettings(s)
}

// GetSettings returns the current settings.
func (p *Provider) GetSettings() *Settings {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.settings
}
