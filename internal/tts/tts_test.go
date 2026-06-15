package tts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockSynthesizer struct {
	synthesizeFunc func(text, voice string) ([]byte, error)
	closeFunc      func() error
	mu             sync.Mutex
	closed         bool
}

func (m *mockSynthesizer) synthesize(text, voice string) ([]byte, error) {
	if m.synthesizeFunc != nil {
		return m.synthesizeFunc(text, voice)
	}
	return []byte("fake-audio"), nil
}

func (m *mockSynthesizer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockSynthesizer) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// ── Helpers ──────────────────────────────────────────────────────────────────────

// newTestService creates a Service with a mock client factory.
// If factory is nil, a default mock that returns "fake-audio" is used.
func newTestService(t *testing.T, factory func() (synthesizer, error)) *Service {
	t.Helper()
	svc, err := NewService(t.TempDir(), "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if factory != nil {
		svc.newClient = factory
	}
	return svc
}

// newTestServiceWithSemaphore creates a Service with a custom semaphore capacity.
// This bypasses NewService to allow control over concurrency for testing.
func newTestServiceWithSemaphore(t *testing.T, semCap int, factory func() (synthesizer, error)) *Service {
	t.Helper()
	return &Service{
		cacheDir:    t.TempDir(),
		defaultLang: "en",
		sem:         make(chan struct{}, semCap),
		newClient:   factory,
	}
}

// seedCache writes a pre-computed cache file for the given word and language.
func seedCache(t *testing.T, svc *Service, word, lang string, data []byte) {
	t.Helper()
	voice := VoiceForLanguage(lang)
	key := cacheName(word, lang, voice)
	path := filepath.Join(svc.cacheDir, key)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
}

// alwaysFailClient returns a factory that always returns an error.
func alwaysFailClient() func() (synthesizer, error) {
	return func() (synthesizer, error) {
		return nil, fmt.Errorf("connection refused")
	}
}

// ── Tests: VoiceForLanguage ────────────────────────────────────────────────────────

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
		{"xx", "en-US-AriaNeural"}, // fallback for unknown language
		{"", "en-US-AriaNeural"},   // fallback for empty language
	}
	for _, tt := range tests {
		got := VoiceForLanguage(tt.lang)
		if got != tt.want {
			t.Errorf("VoiceForLanguage(%q) = %q, want %q", tt.lang, got, tt.want)
		}
	}
}

// ── Tests: NewService ───────────────────────────────────────────────────────────────

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

func TestNewService_CacheDirError(t *testing.T) {
	// Create a temp file, then try to create a cache dir underneath it (should fail)
	tmpFile, err := os.CreateTemp("", "tts-test-")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	badPath := tmpFile.Name() + "/subdir"
	_, err = NewService(badPath, "en")
	if err == nil {
		t.Fatal("expected error when cache dir parent is a file, got nil")
	}
}

func TestNewService_DefaultConcurrencyLimit(t *testing.T) {
	svc, err := NewService(t.TempDir(), "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if cap(svc.sem) != 4 {
		t.Errorf("expected semaphore capacity 4, got %d", cap(svc.sem))
	}
}

func TestNewService_SetsDefaultLang(t *testing.T) {
	svc, err := NewService(t.TempDir(), "fr")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if svc.defaultLang != "fr" {
		t.Errorf("expected defaultLang 'fr', got %q", svc.defaultLang)
	}
}

// ── Tests: Synthesize — Cache Hits ──────────────────────────────────────────────────

func TestSynthesize_CacheHit(t *testing.T) {
	svc := newTestService(t, func() (synthesizer, error) {
		t.Fatal("newClient should not be called on cache hit")
		return nil, nil
	})

	expected := []byte("pre-cached-mp3-data")
	seedCache(t, svc, "hello", "en", expected)

	data, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if string(data) != string(expected) {
		t.Fatalf("got %q, want %q", data, expected)
	}
}

func TestSynthesize_DefaultLanguage(t *testing.T) {
	// Use defaultLang="fr" so empty language resolves to French
	svc, err := NewService(t.TempDir(), "fr")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	svc.newClient = func() (synthesizer, error) {
		t.Fatal("newClient should not be called on cache hit")
		return nil, nil
	}

	expected := []byte("french-audio")
	seedCache(t, svc, "bonjour", "fr", expected)

	// Pass empty language — should default to "fr" (service's defaultLang)
	data, err := svc.Synthesize(context.Background(), "bonjour", "")
	if err != nil {
		t.Fatalf("Synthesize with empty lang: %v", err)
	}
	if string(data) != string(expected) {
		t.Fatalf("got %q, want %q", data, expected)
	}
}

// ── Tests: Synthesize — Cache Miss ──────────────────────────────────────────────────

func TestSynthesize_CacheMiss_CallsClientAndReturnsAudio(t *testing.T) {
	var clientCalls atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		clientCalls.Add(1)
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return []byte("synthesized-audio"), nil
			},
		}, nil
	})

	data, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if string(data) != "synthesized-audio" {
		t.Fatalf("got %q, want %q", data, "synthesized-audio")
	}
	if got := clientCalls.Load(); got != 1 {
		t.Errorf("expected 1 client call, got %d", got)
	}
}

func TestSynthesize_CacheMiss_WritesToCacheForFutureReads(t *testing.T) {
	var clientCalls atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		clientCalls.Add(1)
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return []byte("audio-data"), nil
			},
		}, nil
	})

	// First call — cache miss, calls client
	data1, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if string(data1) != "audio-data" {
		t.Errorf("first call: got %q, want %q", data1, "audio-data")
	}
	if got := clientCalls.Load(); got != 1 {
		t.Errorf("expected 1 client call after first Synthesize, got %d", got)
	}

	// Second call — cache hit, should NOT call client
	data2, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if string(data2) != "audio-data" {
		t.Errorf("second call: got %q, want %q", data2, "audio-data")
	}
	if got := clientCalls.Load(); got != 1 {
		t.Errorf("expected 1 client call total (cached), got %d", got)
	}
}

// ── Tests: Synthesize — Retry Logic ─────────────────────────────────────────────────

func TestSynthesize_RetryOnConnectFailure(t *testing.T) {
	var attempts atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		call := attempts.Add(1)
		if call == 1 {
			return nil, fmt.Errorf("connection refused")
		}
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return []byte("audio-on-retry"), nil
			},
		}, nil
	})

	data, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if string(data) != "audio-on-retry" {
		t.Errorf("got %q, want %q", data, "audio-on-retry")
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts (1 fail + 1 success), got %d", got)
	}
}

func TestSynthesize_RetryOnSynthesizeFailure(t *testing.T) {
	var attempts atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		call := attempts.Add(1)
		if call == 1 {
			return &mockSynthesizer{
				synthesizeFunc: func(text, voice string) ([]byte, error) {
					return nil, fmt.Errorf("synthesize error")
				},
			}, nil
		}
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return []byte("audio-on-retry"), nil
			},
		}, nil
	})

	data, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if string(data) != "audio-on-retry" {
		t.Errorf("got %q, want %q", data, "audio-on-retry")
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts (1 fail + 1 success), got %d", got)
	}
}

func TestSynthesize_BothAttemptsFail(t *testing.T) {
	var attempts atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		attempts.Add(1)
		return nil, fmt.Errorf("connection refused")
	})

	_, err := svc.Synthesize(context.Background(), "hello", "en")
	if err == nil {
		t.Fatal("expected error when all attempts fail, got nil")
	}
	if !strings.Contains(err.Error(), "edge-tts") {
		t.Errorf("expected error to contain 'edge-tts', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected error to contain 'connection refused', got %q", err.Error())
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestSynthesize_SynthesizeFailsBothAttempts(t *testing.T) {
	var attempts atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		attempts.Add(1)
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return nil, fmt.Errorf("permanent error")
			},
		}, nil
	})

	_, err := svc.Synthesize(context.Background(), "hello", "en")
	if err == nil {
		t.Fatal("expected error when both synthesize attempts fail, got nil")
	}
	if !strings.Contains(err.Error(), "edge-tts") {
		t.Errorf("expected error to contain 'edge-tts', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "permanent error") {
		t.Errorf("expected error to contain 'permanent error', got %q", err.Error())
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

// ── Tests: Synthesize — Client Lifecycle ─────────────────────────────────────────────

func TestSynthesize_ClientClosedOnSuccess(t *testing.T) {
	mock := &mockSynthesizer{
		synthesizeFunc: func(text, voice string) ([]byte, error) {
			return []byte("audio"), nil
		},
	}
	svc := newTestService(t, func() (synthesizer, error) {
		return mock, nil
	})

	_, err := svc.Synthesize(context.Background(), "hello", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.isClosed() {
		t.Error("expected client to be closed after successful synthesize")
	}
}

func TestSynthesize_ClientClosedOnSynthesizeFailure(t *testing.T) {
	var clients []*mockSynthesizer
	svc := newTestService(t, func() (synthesizer, error) {
		m := &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return nil, fmt.Errorf("synthesize error")
			},
		}
		clients = append(clients, m)
		return m, nil
	})

	_, err := svc.Synthesize(context.Background(), "hello", "en")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	for i, c := range clients {
		if !c.isClosed() {
			t.Errorf("client %d was not closed after synthesize failure", i)
		}
	}
}

// ── Tests: Synthesize — Context Cancellation ───────────────────────────────────────

func TestSynthesize_CancelledContextOnSemaphore(t *testing.T) {
	// Use a semaphore with capacity 1, then fill it so the next call blocks.
	svc := newTestServiceWithSemaphore(t, 1, func() (synthesizer, error) {
		t.Fatal("newClient should not be called when context is cancelled")
		return nil, nil
	})

	// Fill the only semaphore slot
	svc.sem <- struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := svc.Synthesize(ctx, "hello", "en")
	if err == nil {
		t.Fatal("expected error with cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Drain the slot for cleanup
	<-svc.sem
}

func TestSynthesize_ContextDeadlineOnSemaphore(t *testing.T) {
	// Use a semaphore with capacity 1, then fill it so the next call blocks.
	mockFactory := func() (synthesizer, error) {
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				time.Sleep(200 * time.Millisecond)
				return []byte("audio"), nil
			},
		}, nil
	}
	svc := newTestServiceWithSemaphore(t, 1, mockFactory)

	// Occupies the only slot with a slow operation in the background
	svc.sem <- struct{}{}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := svc.Synthesize(ctx, "hello", "en")
	if err == nil {
		t.Fatal("expected error when context deadline exceeded, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	// Drain the slot for cleanup
	<-svc.sem
}

// ── Tests: Synthesize — Concurrency ────────────────────────────────────────────────

func TestSynthesize_ConcurrentRequests(t *testing.T) {
	var clientCalls atomic.Int32
	svc := newTestService(t, func() (synthesizer, error) {
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				clientCalls.Add(1)
				return []byte("audio-" + text), nil
			},
		}, nil
	})

	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			word := fmt.Sprintf("word-%d", i)
			_, errs[i] = svc.Synthesize(context.Background(), word, "en")
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: unexpected error: %v", i, err)
		}
	}

	if got := clientCalls.Load(); got != n {
		t.Errorf("expected %d client calls (one per unique word), got %d", n, got)
	}
}

// ── Tests: Synthesize — Error Wrapping ─────────────────────────────────────────────

func TestSynthesize_CacheWriteFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping: running as root, permission checks unreliable")
	}

	cacheDir := t.TempDir()
	svc, err := NewService(cacheDir, "en")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	svc.newClient = func() (synthesizer, error) {
		return &mockSynthesizer{
			synthesizeFunc: func(text, voice string) ([]byte, error) {
				return []byte("audio-data"), nil
			},
		}, nil
	}

	// Make cache dir read-only so the write fails
	if err := os.Chmod(cacheDir, 0550); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(cacheDir, 0750) // restore for cleanup

	// Verify the dir is actually read-only
	f, err := os.CreateTemp(cacheDir, "write-check-")
	if err == nil {
		f.Close()
		os.Remove(f.Name())
		t.Skip("skipping: could not make cache dir read-only")
	}

	_, err = svc.Synthesize(context.Background(), "hello", "en")
	if err == nil {
		t.Fatal("expected error when cache write fails, got nil")
	}
	if !strings.Contains(err.Error(), "cache write") {
		t.Errorf("expected error to contain 'cache write', got %q", err.Error())
	}
}

// ── Tests: cacheName ──────────────────────────────────────────────────────────────

func TestCacheName_Deterministic(t *testing.T) {
	a := cacheName("hello", "en", "en-US-AriaNeural")
	b := cacheName("hello", "en", "en-US-AriaNeural")
	if a != b {
		t.Fatalf("cache names should be deterministic: %q != %q", a, b)
	}
}

func TestCacheName_DifferentWord(t *testing.T) {
	a := cacheName("hello", "en", "en-US-AriaNeural")
	b := cacheName("world", "en", "en-US-AriaNeural")
	if a == b {
		t.Fatal("different words should produce different cache names")
	}
}

func TestCacheName_DifferentLanguage(t *testing.T) {
	a := cacheName("hello", "en", "en-US-AriaNeural")
	b := cacheName("hello", "es", "es-ES-ElviraNeural")
	if a == b {
		t.Fatal("different languages should produce different cache names")
	}
}

func TestCacheName_FormatHasMP3Suffix(t *testing.T) {
	name := cacheName("hello", "en", "en-US-AriaNeural")
	if !strings.HasSuffix(name, ".mp3") {
		t.Errorf("cache name should end with .mp3, got %q", name)
	}
}

// ── Tests: escapeXML ──────────────────────────────────────────────────────────────

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no special chars", "hello world", "hello world"},
		{"ampersand", "rock & roll", "rock &amp; roll"},
		{"less than", "a < b", "a &lt; b"},
		{"greater than", "a > b", "a &gt; b"},
		{"apostrophe", "it's", "it&apos;s"},
		{"quote", `say "hello"`, "say &quot;hello&quot;"},
		{"all special chars", `&<>'"`, "&amp;&lt;&gt;&apos;&quot;"},
		{"ampersand before lt", "&lt;", "&amp;lt;"},
		{"mixed", `He said "a < b & c > d"`, "He said &quot;a &lt; b &amp; c &gt; d&quot;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeXML(tt.input)
			if got != tt.want {
				t.Errorf("escapeXML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
