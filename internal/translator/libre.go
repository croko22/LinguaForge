package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LibreClient struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

// NewLibreTranslate creates a Translator backed by a LibreTranslate-compatible API.
func NewLibreTranslate(endpoint string, apiKey string) Translator {
	return &LibreClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

type libreRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
	APIKey string `json:"api_key,omitempty"`
}

type libreResponse struct {
	TranslatedText string `json:"translatedText"`
}

func (l *LibreClient) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
	if req.Word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	body := libreRequest{
		Q:      req.Word,
		Source: req.SourceLang,
		Target: req.TargetLang,
	}
	if l.apiKey != "" {
		body.APIKey = l.apiKey
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, l.endpoint+"/translate", &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("libre translate returned status %d", resp.StatusCode)
	}

	var libreResp libreResponse
	if err := json.NewDecoder(resp.Body).Decode(&libreResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &TranslateResponse{Translation: libreResp.TranslatedText}, nil
}
