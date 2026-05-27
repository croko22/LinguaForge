package translator

import (
	"context"
	"testing"
)

func TestMockTranslator_Translate(t *testing.T) {
	m := NewMockTranslator()

	result, err := m.Translate(context.Background(), TranslateRequest{
		Word:       "hello",
		SourceLang: "en",
		TargetLang: "es",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Translation == "" {
		t.Error("expected non-empty translation")
	}
}

func TestMockTranslator_Translate_EmptyWord(t *testing.T) {
	m := NewMockTranslator()

	_, err := m.Translate(context.Background(), TranslateRequest{
		Word:       "",
		SourceLang: "en",
		TargetLang: "es",
	})
	if err == nil {
		t.Error("expected error for empty word")
	}
}
