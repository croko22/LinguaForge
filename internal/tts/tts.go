package tts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

var voiceForLang = map[string]string{
	"en": "en-US-AriaNeural",
	"es": "es-ES-ElviraNeural",
	"fr": "fr-FR-DeniseNeural",
	"de": "de-DE-KatjaNeural",
	"it": "it-IT-ElsaNeural",
	"pt": "pt-BR-FranciscaNeural",
	"ja": "ja-JP-NanamiNeural",
	"ko": "ko-KR-SunHiNeural",
	"zh": "zh-CN-XiaoxiaoNeural",
	"ru": "ru-RU-SvetlanaNeural",
	"ar": "ar-SA-ZariyahNeural",
	"nl": "nl-NL-MaartenNeural",
	"pl": "pl-PL-AgnieszkaNeural",
	"sv": "sv-SE-SofieNeural",
	"da": "da-DK-ChristelNeural",
	"fi": "fi-FI-SelmaNeural",
	"nb": "nb-NO-PernilleNeural",
	"tr": "tr-TR-EmelNeural",
}

func VoiceForLanguage(lang string) string {
	if v, ok := voiceForLang[lang]; ok {
		return v
	}
	return "en-US-AriaNeural"
}

// synthesizer abstracts the TTS client for testability.
type synthesizer interface {
	synthesize(text, voice string) ([]byte, error)
	Close() error
}

type Service struct {
	cacheDir    string
	defaultLang string
	sem         chan struct{}
	newClient   func() (synthesizer, error)
}

// defaultNewClient is the production factory that creates a real edge-tts client.
var defaultNewClient = func() (synthesizer, error) {
	return newEdgeClient()
}

func NewService(cacheDir, defaultLang string) (*Service, error) {
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
		return nil, fmt.Errorf("create tts cache dir: %w", err)
	}

	return &Service{
		cacheDir:    cacheDir,
		defaultLang: defaultLang,
		sem:         make(chan struct{}, 4),
		newClient:   defaultNewClient,
	}, nil
}

func (s *Service) Synthesize(ctx context.Context, word, language string) ([]byte, error) {
	if language == "" {
		language = s.defaultLang
	}

	voice := VoiceForLanguage(language)
	cacheKey := cacheName(word, language, voice)
	cachePath := filepath.Join(s.cacheDir, cacheKey)

	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-s.sem }()

	// Try twice — edge-tts WebSocket can flake
	var data []byte
	var err error
	for attempt := range 2 {
		var client synthesizer
		client, err = s.newClient()
		if err != nil {
			slog.Warn("tts: edge connect failed", "attempt", attempt+1, "error", err)
			continue
		}
		data, err = client.synthesize(word, voice)
		client.Close()
		if err == nil {
			break
		}
		slog.Warn("tts: edge synthesize failed", "attempt", attempt+1, "word", word, "error", err)
	}
	if err != nil {
		return nil, fmt.Errorf("edge-tts: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return nil, fmt.Errorf("cache write: %w", err)
	}

	return data, nil
}

func cacheName(word, language, voice string) string {
	h := sha256.Sum256([]byte(word + "|" + language + "|" + voice))
	return hex.EncodeToString(h[:]) + ".mp3"
}
