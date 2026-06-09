package tts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestVoiceForLanguage(t *testing.T) {
	tests := []struct {
		lang string
		want string
	}{
		{"en", "en-US-AriaNeural"},
		{"es", "es-ES-ElviraNeural"},
		{"fr", "fr-FR-DeniseNeural"},
		{"de", "de-DE-KatjaNeural"},
		{"it", "it-IT-ElsaNeural"},
		{"pt", "pt-BR-FranciscaNeural"},
		{"ja", "ja-JP-NanamiNeural"},
		{"ko", "ko-KR-SunHiNeural"},
		{"zh", "zh-CN-XiaoxiaoNeural"},
		{"ru", "ru-RU-SvetlanaNeural"},
		{"ar", "ar-SA-ZariyahNeural"},
		{"nl", "nl-NL-MaartenNeural"},
		{"pl", "pl-PL-AgnieszkaNeural"},
		{"sv", "sv-SE-SofieNeural"},
		{"da", "da-DK-ChristelNeural"},
		{"fi", "fi-FI-SelmaNeural"},
		{"nb", "nb-NO-PernilleNeural"},
		{"tr", "tr-TR-EmelNeural"},
		{"xx", "en-US-AriaNeural"}, // fallback
		{"", "en-US-AriaNeural"},   // fallback
	}
	for _, tt := range tests {
		got := VoiceForLanguage(tt.lang)
		if got != tt.want {
			t.Errorf("VoiceForLanguage(%q) = %q, want %q", tt.lang, got, tt.want)
		}
	}
}

func TestNewService_CreatesCacheDir(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "tts-cache")

	svc, err := NewService(cacheDir, "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatal("cache directory was not created")
	}
}

func TestSynthesize_CacheHit(t *testing.T) {
	cacheDir := t.TempDir()
	svc, err := NewService(cacheDir, "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	word := "hello"
	lang := "en"
	voice := VoiceForLanguage(lang)

	cacheName := cacheNameForTest(word, lang, voice)
	cachePath := filepath.Join(cacheDir, cacheName)
	expected := []byte("pre-cached-mp3-data")
	if err := os.WriteFile(cachePath, expected, 0644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	data, err := svc.Synthesize(context.Background(), word, lang)
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}

	if string(data) != string(expected) {
		t.Fatalf("got %q, want %q", data, expected)
	}
}

func TestSynthesize_DefaultLanguage(t *testing.T) {
	cacheDir := t.TempDir()
	svc, err := NewService(cacheDir, "fr")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	word := "bonjour"
	voice := VoiceForLanguage("fr")
	cacheName := cacheNameForTest(word, "fr", voice)
	cachePath := filepath.Join(cacheDir, cacheName)
	expected := []byte("french-audio")
	if err := os.WriteFile(cachePath, expected, 0644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	data, err := svc.Synthesize(context.Background(), word, "")
	if err != nil {
		t.Fatalf("Synthesize with empty lang: %v", err)
	}

	if string(data) != string(expected) {
		t.Fatalf("got %q, want %q", data, expected)
	}
}

func TestSynthesize_DownstreamFails(t *testing.T) {
	cacheDir := t.TempDir()
	svc, err := NewService(cacheDir, "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Synthesize(context.Background(), "hello", "en")
	if err == nil {
		t.Fatal("expected error when edge-tts is unavailable, got nil")
	}
}

func TestCacheName_Deterministic(t *testing.T) {
	a := cacheNameForTest("hello", "en", "en-US-AriaNeural")
	b := cacheNameForTest("hello", "en", "en-US-AriaNeural")
	if a != b {
		t.Fatalf("cache names should be deterministic: %q != %q", a, b)
	}
}

func TestCacheName_DifferentWord(t *testing.T) {
	a := cacheNameForTest("hello", "en", "en-US-AriaNeural")
	b := cacheNameForTest("world", "en", "en-US-AriaNeural")
	if a == b {
		t.Fatal("different words should produce different cache names")
	}
}

func cacheNameForTest(word, language, voice string) string {
	h := sha256.Sum256([]byte(word + "|" + language + "|" + voice))
	return hex.EncodeToString(h[:]) + ".mp3"
}
