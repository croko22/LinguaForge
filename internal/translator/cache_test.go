package translator

import (
	"context"
	"sync"
	"testing"
)

func TestCachedTranslator_CachesResult(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	mock := NewMockTranslator()
	counting := &callCounter{inner: mock, count: &callCount, mu: &mu}
	cached := NewCachedTranslator(counting)

	// First call
	_, err := cached.Translate(context.Background(), TranslateRequest{Word: "hello", SourceLang: "en", TargetLang: "es"})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Second call with same word
	_, err = cached.Translate(context.Background(), TranslateRequest{Word: "hello", SourceLang: "en", TargetLang: "es"})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	mu.Lock()
	if callCount != 1 {
		t.Fatalf("expected 1 call to inner translator, got %d", callCount)
	}
	mu.Unlock()
}

func TestCachedTranslator_DifferentWords(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	mock := NewMockTranslator()
	counting := &callCounter{inner: mock, count: &callCount, mu: &mu}
	cached := NewCachedTranslator(counting)

	cached.Translate(context.Background(), TranslateRequest{Word: "hello", SourceLang: "en", TargetLang: "es"})
	cached.Translate(context.Background(), TranslateRequest{Word: "world", SourceLang: "en", TargetLang: "es"})

	mu.Lock()
	if callCount != 2 {
		t.Fatalf("expected 2 calls for different words, got %d", callCount)
	}
	mu.Unlock()
}

func TestCachedTranslator_EmptyWord(t *testing.T) {
	cached := NewCachedTranslator(NewMockTranslator())
	_, err := cached.Translate(context.Background(), TranslateRequest{Word: "", SourceLang: "en", TargetLang: "es"})
	if err == nil {
		t.Fatal("expected error for empty word")
	}
}
