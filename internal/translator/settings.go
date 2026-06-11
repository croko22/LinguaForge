package translator

// ProviderConfig holds connection details for a translation provider.
type ProviderConfig struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key,omitempty"`
}

// Settings holds the current translation configuration.
type Settings struct {
	ActiveProvider string           `json:"active_provider"`
	Providers      []ProviderConfig `json:"providers"`
}

// DefaultSettings returns sensible defaults.
func DefaultSettings() *Settings {
	return &Settings{
		ActiveProvider: "mock",
		Providers: []ProviderConfig{
			{Name: "mock", Endpoint: "", APIKey: ""},
			{Name: "libre", Endpoint: "https://libretranslate.com", APIKey: ""},
		},
	}
}

// WithLibreAsDefault returns settings with libre as the active provider.
// Used when a local LibreTranslate endpoint is configured (e.g. Docker).
func WithLibreAsDefault(endpoint, apiKey string) *Settings {
	return &Settings{
		ActiveProvider: "libre",
		Providers: []ProviderConfig{
			{Name: "mock", Endpoint: "", APIKey: ""},
			{Name: "libre", Endpoint: endpoint, APIKey: apiKey},
		},
	}
}
