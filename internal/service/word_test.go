package service

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/translator"
)

// ── Mock Translator ────────────────────────────────────────────────────────────

type mockTranslator struct {
	translateFunc func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error)
}

func (m *mockTranslator) Translate(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
	if m.translateFunc != nil {
		return m.translateFunc(ctx, req)
	}
	return nil, fmt.Errorf("translateFunc not set")
}

// ── Helpers ──────────────────────────────────────────────────────────────────────

func newWordService(wordRepo *mockWordRepo, reviewRepo *mockReviewRepo, t translator.Translator) *WordService {
	return NewWordService(wordRepo, reviewRepo, t)
}

// ── Tests: SaveWord ──────────────────────────────────────────────────────────────

func TestSaveWord_WithExplicitTranslation_SkipsTranslator(t *testing.T) {
	var translatorCalls int32
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			atomic.AddInt32(&translatorCalls, 1)
			return &translator.TranslateResponse{Translation: "should not be called"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	result, err := svc.SaveWord(context.Background(), "doc-1", "hello", "hola", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Translation != "hola" {
		t.Errorf("expected translation 'hola', got %q", result.Translation)
	}
	if result.Word != "hello" {
		t.Errorf("expected word 'hello', got %q", result.Word)
	}
	if result.SourceLang != "en" {
		t.Errorf("expected source_lang 'en', got %q", result.SourceLang)
	}
	if result.TargetLang != "es" {
		t.Errorf("expected target_lang 'es', got %q", result.TargetLang)
	}
	if result.DocumentID != "doc-1" {
		t.Errorf("expected document_id 'doc-1', got %q", result.DocumentID)
	}
	if calls := atomic.LoadInt32(&translatorCalls); calls != 0 {
		t.Errorf("expected 0 translator calls, got %d", calls)
	}
}

func TestSaveWord_EmptyTranslation_CallsTranslatorOnceAndSucceeds(t *testing.T) {
	var translatorCalls int32
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			atomic.AddInt32(&translatorCalls, 1)
			if req.Word != "hello" {
				t.Errorf("expected word 'hello', got %q", req.Word)
			}
			if req.SourceLang != "en" {
				t.Errorf("expected source_lang 'en', got %q", req.SourceLang)
			}
			if req.TargetLang != "es" {
				t.Errorf("expected target_lang 'es', got %q", req.TargetLang)
			}
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	result, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Translation != "hola" {
		t.Errorf("expected translation 'hola', got %q", result.Translation)
	}
	if calls := atomic.LoadInt32(&translatorCalls); calls != 1 {
		t.Errorf("expected 1 translator call, got %d", calls)
	}
}

func TestSaveWord_EmptyTranslation_TranslatorFailsOnce_SucceedsOnRetry(t *testing.T) {
	var translatorCalls int32
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			call := atomic.AddInt32(&translatorCalls, 1)
			if call == 1 {
				return nil, fmt.Errorf("transient error")
			}
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	result, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Translation != "hola" {
		t.Errorf("expected translation 'hola', got %q", result.Translation)
	}
	if calls := atomic.LoadInt32(&translatorCalls); calls != 2 {
		t.Errorf("expected 2 translator calls, got %d", calls)
	}
}

func TestSaveWord_EmptyTranslation_TranslatorFailsTwice_SucceedsOnThirdAttempt(t *testing.T) {
	var translatorCalls int32
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			call := atomic.AddInt32(&translatorCalls, 1)
			if call <= 2 {
				return nil, fmt.Errorf("transient error attempt %d", call)
			}
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	result, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Translation != "hola" {
		t.Errorf("expected translation 'hola', got %q", result.Translation)
	}
	if calls := atomic.LoadInt32(&translatorCalls); calls != 3 {
		t.Errorf("expected 3 translator calls, got %d", calls)
	}
}

func TestSaveWord_EmptyTranslation_TranslatorFailsAllAttempts_ReturnsError(t *testing.T) {
	var translatorCalls int32
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			atomic.AddInt32(&translatorCalls, 1)
			return nil, fmt.Errorf("permanent error")
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify the error message includes retry count and word
	expectedSubstring := "translate word \"hello\" after 3 attempts"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("expected error to contain %q, got %q", expectedSubstring, err.Error())
	}
	if !strings.Contains(err.Error(), "permanent error") {
		t.Errorf("expected error to contain 'permanent error', got %q", err.Error())
	}
	if calls := atomic.LoadInt32(&translatorCalls); calls != 3 {
		t.Errorf("expected 3 translator calls, got %d", calls)
	}
}

func TestSaveWord_ContextCancelledDuringBackoff_ReturnsContextError(t *testing.T) {
	var translatorCalls int32
	ctx, cancel := context.WithCancel(context.Background())

	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			call := atomic.AddInt32(&translatorCalls, 1)
			if call == 1 {
				// Cancel context after first failed attempt to trigger cancellation during backoff
				cancel()
				return nil, fmt.Errorf("transient error")
			}
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.SaveWord(ctx, "doc-1", "hello", "", "en", "es")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// The error should be a context error
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSaveWord_ContextCancelledDuringTranslatorCall(t *testing.T) {
	// Test that a context already cancelled before SaveWord returns immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			return nil, fmt.Errorf("should not be called")
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.SaveWord(ctx, "doc-1", "hello", "", "en", "es")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// The translator may or may not be called depending on whether the context
	// cancellation propagates through the Translate call. The key point is
	// the error should be a context error.
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSaveWord_AutoCreatesReviewCard(t *testing.T) {
	var capturedCard *model.ReviewCard
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			capturedCard = card
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	result, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedCard == nil {
		t.Fatal("expected review card to be created, got nil")
	}
	if capturedCard.WordID != result.ID {
		t.Errorf("expected card WordID=%s, got %s", result.ID, capturedCard.WordID)
	}
	if capturedCard.Status != model.ReviewStatusNew {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusNew, capturedCard.Status)
	}
	if capturedCard.EaseFactor != 2.5 {
		t.Errorf("expected ease_factor=2.5, got %f", capturedCard.EaseFactor)
	}
	if capturedCard.IntervalDays != 0 {
		t.Errorf("expected interval_days=0, got %d", capturedCard.IntervalDays)
	}
	if capturedCard.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", capturedCard.Repetitions)
	}
	if capturedCard.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", capturedCard.Lapses)
	}
	if capturedCard.ID == "" {
		t.Error("expected non-empty card ID")
	}

	// Verify next_review is approximately now
	now := time.Now().UTC()
	diff := now.Sub(capturedCard.NextReview)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("expected next_review ≈ now, got diff=%v", diff)
	}
}

func TestSaveWord_ReviewCardFailure_DoesNotFailWordSave(t *testing.T) {
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return fmt.Errorf("review card creation failed")
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	result, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err != nil {
		t.Fatalf("word save should not fail when review card creation fails, got error: %v", err)
	}

	// The word should still be returned with its data
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Translation != "hola" {
		t.Errorf("expected translation 'hola', got %q", result.Translation)
	}
	if result.Word != "hello" {
		t.Errorf("expected word 'hello', got %q", result.Word)
	}
}

func TestSaveWord_RepoSaveFailure_ReturnsError(t *testing.T) {
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			return &translator.TranslateResponse{Translation: "hola"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			return fmt.Errorf("db connection lost")
		},
	}
	reviewRepo := &mockReviewRepo{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.SaveWord(context.Background(), "doc-1", "hello", "", "en", "es")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "save word") {
		t.Errorf("expected error to contain 'save word', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "db connection lost") {
		t.Errorf("expected error to contain 'db connection lost', got %q", err.Error())
	}
}

func TestSaveWord_SetsFieldsCorrectly(t *testing.T) {
	var capturedWord *model.SavedWord
	mockT := &mockTranslator{
		translateFunc: func(ctx context.Context, req translator.TranslateRequest) (*translator.TranslateResponse, error) {
			return &translator.TranslateResponse{Translation: "translation"}, nil
		},
	}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			capturedWord = w
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.SaveWord(context.Background(), "doc-1", "word", "", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedWord == nil {
		t.Fatal("expected word to be saved, got nil")
	}
	if capturedWord.ID == "" {
		t.Error("expected non-empty ID")
	}
	if capturedWord.DocumentID != "doc-1" {
		t.Errorf("expected document_id 'doc-1', got %q", capturedWord.DocumentID)
	}
	if capturedWord.Word != "word" {
		t.Errorf("expected word 'word', got %q", capturedWord.Word)
	}
	if capturedWord.Translation != "translation" {
		t.Errorf("expected translation 'translation', got %q", capturedWord.Translation)
	}
	if capturedWord.SourceLang != "en" {
		t.Errorf("expected source_lang 'en', got %q", capturedWord.SourceLang)
	}
	if capturedWord.TargetLang != "es" {
		t.Errorf("expected target_lang 'es', got %q", capturedWord.TargetLang)
	}
	// Verify CreatedAt is approximately now
	now := time.Now().UTC()
	diff := now.Sub(capturedWord.CreatedAt)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("expected created_at ≈ now, got diff=%v", diff)
	}
}

func TestSaveWord_WithExplicitTranslation_SetsProvidedTranslation(t *testing.T) {
	var capturedWord *model.SavedWord
	mockT := &mockTranslator{}

	wordRepo := &mockWordRepo{
		saveFunc: func(ctx context.Context, w *model.SavedWord) error {
			capturedWord = w
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return nil
		},
	}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.SaveWord(context.Background(), "doc-1", "hello", "custom-translation", "fr", "de")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedWord.Translation != "custom-translation" {
		t.Errorf("expected translation 'custom-translation', got %q", capturedWord.Translation)
	}
	if capturedWord.SourceLang != "fr" {
		t.Errorf("expected source_lang 'fr', got %q", capturedWord.SourceLang)
	}
	if capturedWord.TargetLang != "de" {
		t.Errorf("expected target_lang 'de', got %q", capturedWord.TargetLang)
	}
}

// ── Tests: ListWords ────────────────────────────────────────────────────────────

func TestListWords_DelegatesToRepo(t *testing.T) {
	expected := []*model.SavedWord{
		{ID: "w1", Word: "hello", Translation: "hola", SourceLang: "en", TargetLang: "es", DocumentID: "doc-1"},
		{ID: "w2", Word: "book", Translation: "libro", SourceLang: "en", TargetLang: "es", DocumentID: "doc-1"},
	}

	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return expected, nil
		},
	}
	reviewRepo := &mockReviewRepo{}
	mockT := &mockTranslator{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	words, err := svc.ListWords(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
	if words[0].ID != "w1" {
		t.Errorf("expected ID 'w1', got %q", words[0].ID)
	}
	if words[1].ID != "w2" {
		t.Errorf("expected ID 'w2', got %q", words[1].ID)
	}
}

func TestListWords_EmptyResult(t *testing.T) {
	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return []*model.SavedWord{}, nil
		},
	}
	reviewRepo := &mockReviewRepo{}
	mockT := &mockTranslator{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	words, err := svc.ListWords(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if words == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words, got %d", len(words))
	}
}

func TestListWords_RepoError_ReturnsError(t *testing.T) {
	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return nil, fmt.Errorf("db connection lost")
		},
	}
	reviewRepo := &mockReviewRepo{}
	mockT := &mockTranslator{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	_, err := svc.ListWords(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "db connection lost") {
		t.Errorf("expected error containing 'db connection lost', got %q", err.Error())
	}
}

// ── Tests: DeleteWord ────────────────────────────────────────────────────────────

func TestDeleteWord_DelegatesToRepo(t *testing.T) {
	var deletedID string
	wordRepo := &mockWordRepo{
		deleteFunc: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
	}
	reviewRepo := &mockReviewRepo{}
	mockT := &mockTranslator{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	err := svc.DeleteWord(context.Background(), "word-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != "word-123" {
		t.Errorf("expected deleted ID 'word-123', got %q", deletedID)
	}
}

func TestDeleteWord_RepoError_ReturnsError(t *testing.T) {
	wordRepo := &mockWordRepo{
		deleteFunc: func(ctx context.Context, id string) error {
			return fmt.Errorf("word not found")
		},
	}
	reviewRepo := &mockReviewRepo{}
	mockT := &mockTranslator{}

	svc := newWordService(wordRepo, reviewRepo, mockT)
	err := svc.DeleteWord(context.Background(), "word-nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "word not found") {
		t.Errorf("expected error containing 'word not found', got %q", err.Error())
	}
}
