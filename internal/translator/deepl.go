package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DeepLClient struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewDeepLTranslate(endpoint, apiKey string) Translator {
	return &DeepLClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

type deeplRequest struct {
	Text       []string `json:"text"`
	SourceLang string   `json:"source_lang"`
	TargetLang string   `json:"target_lang"`
}

type deeplResponse struct {
	Translations []deeplTranslation `json:"translations"`
}

type deeplTranslation struct {
	Text string `json:"text"`
}

func (d *DeepLClient) Translate(ctx context.Context, req TranslateRequest) (*TranslateResponse, error) {
	if req.Word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	body := deeplRequest{
		Text:       []string{req.Word},
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint+"/v2/translate", &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "DeepL-Auth-Key "+d.apiKey)

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepl translate returned status %d", resp.StatusCode)
	}

	var deeplResp deeplResponse
	if err := json.NewDecoder(resp.Body).Decode(&deeplResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(deeplResp.Translations) == 0 {
		return nil, fmt.Errorf("deepl returned empty translations")
	}

	return &TranslateResponse{Translation: deeplResp.Translations[0].Text}, nil
}
