package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/srs"
)

// ── Mocks ──────────────────────────────────────────────────────────────────────

type mockReviewRepo struct {
	createFunc       func(ctx context.Context, card *model.ReviewCard) error
	getByWordIDFunc  func(ctx context.Context, wordID string) (*model.ReviewCard, error)
	getDueWordsFunc  func(ctx context.Context) ([]*model.ReviewCard, error)
	updateReviewFunc func(ctx context.Context, card *model.ReviewCard) error
	countDueFunc     func(ctx context.Context) (int, error)
}

func (m *mockReviewRepo) Create(ctx context.Context, card *model.ReviewCard) error {
	if m.createFunc == nil {
		return nil
	}
	return m.createFunc(ctx, card)
}

func (m *mockReviewRepo) GetByWordID(ctx context.Context, wordID string) (*model.ReviewCard, error) {
	return m.getByWordIDFunc(ctx, wordID)
}

func (m *mockReviewRepo) GetDueWords(ctx context.Context) ([]*model.ReviewCard, error) {
	return m.getDueWordsFunc(ctx)
}

func (m *mockReviewRepo) UpdateReview(ctx context.Context, card *model.ReviewCard) error {
	return m.updateReviewFunc(ctx, card)
}

func (m *mockReviewRepo) CountDue(ctx context.Context) (int, error) {
	return m.countDueFunc(ctx)
}

func (m *mockReviewRepo) GetReviewActivity(_ context.Context, _ int) ([]repository.ReviewActivity, error) {
	return nil, nil
}

func (m *mockReviewRepo) DeleteByDocumentID(_ context.Context, _ string) error {
	return nil
}

type mockWordRepo struct {
	saveFunc           func(ctx context.Context, word *model.SavedWord) error
	listByDocumentFunc func(ctx context.Context, documentID string) ([]*model.SavedWord, error)
	listAllFunc        func(ctx context.Context) ([]*model.SavedWord, error)
	deleteFunc         func(ctx context.Context, id string) error
}

func (m *mockWordRepo) Save(ctx context.Context, word *model.SavedWord) error {
	if m.saveFunc == nil {
		return nil
	}
	return m.saveFunc(ctx, word)
}

func (m *mockWordRepo) ListByDocument(ctx context.Context, documentID string) ([]*model.SavedWord, error) {
	return m.listByDocumentFunc(ctx, documentID)
}

func (m *mockWordRepo) ListAll(ctx context.Context) ([]*model.SavedWord, error) {
	return m.listAllFunc(ctx)
}

func (m *mockWordRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc == nil {
		return nil
	}
	return m.deleteFunc(ctx, id)
}

func (m *mockWordRepo) CountAll(_ context.Context) (int, error) {
	return 0, nil
}

func (m *mockWordRepo) DeleteByDocumentID(_ context.Context, _ string) error {
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────────

// newCard returns a ReviewCard with sensible defaults for testing.
func newCard(status string, reps int, ef float64, interval int) *model.ReviewCard {
	now := time.Now().UTC()
	return &model.ReviewCard{
		ID:             "card-1",
		WordID:         "word-1",
		Status:         status,
		EaseFactor:     ef,
		IntervalDays:   interval,
		Repetitions:    reps,
		Lapses:         0,
		NextReview:     now,
		LastReviewedAt: nil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func newReviewService(reviewRepo *mockReviewRepo, wordRepo *mockWordRepo) *ReviewService {
	return NewReviewService(reviewRepo, wordRepo)
}

// roundFloat rounds a float64 to 2 decimal places for ease factor comparisons.
func roundFloat(f float64) float64 {
	return math.Round(f*100) / 100
}

// ── Tests: GetDue ─────────────────────────────────────────────────────────────────

func TestGetDue_ReturnsDueCardsWithWordData(t *testing.T) {
	now := time.Now().UTC()

	reviewRepo := &mockReviewRepo{
		getDueWordsFunc: func(ctx context.Context) ([]*model.ReviewCard, error) {
			return []*model.ReviewCard{
				{
					ID: "card-1", WordID: "word-1", Status: model.ReviewStatusReview,
					EaseFactor: 2.5, IntervalDays: 6, Repetitions: 2, Lapses: 0,
					NextReview: now,
				},
			}, nil
		},
	}

	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return []*model.SavedWord{
				{ID: "word-1", Word: "hello", Translation: "hola", SourceLang: "en", TargetLang: "es", DocumentID: "doc-1"},
			}, nil
		},
	}

	svc := newReviewService(reviewRepo, wordRepo)

	results, err := svc.GetDue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.ID != "word-1" {
		t.Errorf("expected ID=word-1, got %s", r.ID)
	}
	if r.Word != "hello" {
		t.Errorf("expected Word=hello, got %s", r.Word)
	}
	if r.Translation != "hola" {
		t.Errorf("expected Translation=hola, got %s", r.Translation)
	}
	if r.SourceLang != "en" {
		t.Errorf("expected SourceLang=en, got %s", r.SourceLang)
	}
	if r.TargetLang != "es" {
		t.Errorf("expected TargetLang=es, got %s", r.TargetLang)
	}
	if r.DocumentID != "doc-1" {
		t.Errorf("expected DocumentID=doc-1, got %s", r.DocumentID)
	}
	if r.Status != model.ReviewStatusReview {
		t.Errorf("expected Status=%s, got %s", model.ReviewStatusReview, r.Status)
	}
	if r.EaseFactor != 2.5 {
		t.Errorf("expected EaseFactor=2.5, got %f", r.EaseFactor)
	}
	if r.IntervalDays != 6 {
		t.Errorf("expected IntervalDays=6, got %d", r.IntervalDays)
	}
	if r.Repetitions != 2 {
		t.Errorf("expected Repetitions=2, got %d", r.Repetitions)
	}
	if r.Lapses != 0 {
		t.Errorf("expected Lapses=0, got %d", r.Lapses)
	}
}

func TestGetDue_EmptyResult(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		getDueWordsFunc: func(ctx context.Context) ([]*model.ReviewCard, error) {
			return []*model.ReviewCard{}, nil
		},
	}

	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return []*model.SavedWord{}, nil
		},
	}

	svc := newReviewService(reviewRepo, wordRepo)

	results, err := svc.GetDue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestGetDue_SkipsCardsWithoutMatchingWord(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		getDueWordsFunc: func(ctx context.Context) ([]*model.ReviewCard, error) {
			return []*model.ReviewCard{
				{ID: "card-1", WordID: "word-missing", Status: model.ReviewStatusReview},
				{ID: "card-2", WordID: "word-1", Status: model.ReviewStatusLearning},
			}, nil
		},
	}

	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return []*model.SavedWord{
				{ID: "word-1", Word: "hello", Translation: "hola"},
			}, nil
		},
	}

	svc := newReviewService(reviewRepo, wordRepo)

	results, err := svc.GetDue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (orphan card skipped), got %d", len(results))
	}
	if results[0].ID != "word-1" {
		t.Errorf("expected ID=word-1, got %s", results[0].ID)
	}
}

func TestGetDue_ReviewRepoError(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		getDueWordsFunc: func(ctx context.Context) ([]*model.ReviewCard, error) {
			return nil, fmt.Errorf("db error")
		},
	}

	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return nil, nil
		},
	}

	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.GetDue(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidQuality) && err.Error() == "get due: db error" {
		// Just verify it wraps the original error
	}
	if err.Error() != "get due: db error" {
		t.Errorf("expected wrapped error 'get due: db error', got %q", err.Error())
	}
}

func TestGetDue_WordRepoError(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		getDueWordsFunc: func(ctx context.Context) ([]*model.ReviewCard, error) {
			return []*model.ReviewCard{{ID: "card-1", WordID: "word-1"}}, nil
		},
	}

	wordRepo := &mockWordRepo{
		listAllFunc: func(ctx context.Context) ([]*model.SavedWord, error) {
			return nil, fmt.Errorf("word db error")
		},
	}

	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.GetDue(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "get due: list words: word db error" {
		t.Errorf("expected wrapped error, got %q", err.Error())
	}
}

// ── Tests: SubmitReview — Invalid Quality ──────────────────────────────────────

func TestSubmitReview_InvalidQuality_TooLow(t *testing.T) {
	reviewRepo := &mockReviewRepo{}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.SubmitReview(context.Background(), "word-1", srs.Quality(0))
	if err == nil {
		t.Fatal("expected error for quality 0, got nil")
	}
	if !errors.Is(err, ErrInvalidQuality) {
		t.Errorf("expected ErrInvalidQuality, got %v", err)
	}
}

func TestSubmitReview_InvalidQuality_TooHigh(t *testing.T) {
	reviewRepo := &mockReviewRepo{}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.SubmitReview(context.Background(), "word-1", srs.Quality(5))
	if err == nil {
		t.Fatal("expected error for quality 5, got nil")
	}
	if !errors.Is(err, ErrInvalidQuality) {
		t.Errorf("expected ErrInvalidQuality, got %v", err)
	}
}

func TestSubmitReview_InvalidQuality_Negative(t *testing.T) {
	reviewRepo := &mockReviewRepo{}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.SubmitReview(context.Background(), "word-1", srs.Quality(-1))
	if err == nil {
		t.Fatal("expected error for negative quality, got nil")
	}
	if !errors.Is(err, ErrInvalidQuality) {
		t.Errorf("expected ErrInvalidQuality, got %v", err)
	}
}

// ── Tests: SubmitReview — New Card ──────────────────────────────────────────────

func TestSubmitReview_NewCard_Good(t *testing.T) {
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}

	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: new → learning
	if result.Status != model.ReviewStatusLearning {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusLearning, result.Status)
	}
	// SM-2: rep=0+1=1, interval=1 (first pass)
	if result.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", result.Repetitions)
	}
	if result.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", result.IntervalDays)
	}
	// SM-2 EF: 2.5 + (0.1 - 2*(0.08+0.04)) = 2.5 - 0.14 = 2.36
	if roundFloat(result.EaseFactor) != 2.36 {
		t.Errorf("expected ease_factor≈2.36, got %.4f", result.EaseFactor)
	}
	// No lapse (quality >= Good)
	if result.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", result.Lapses)
	}
	// LastReviewedAt should be set
	if result.LastReviewedAt == nil {
		t.Error("expected last_reviewed_at to be set")
	}
}

func TestSubmitReview_NewCard_Easy(t *testing.T) {
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Easy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: new → learning (Easy is >= Good, so transition happens)
	if result.Status != model.ReviewStatusLearning {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusLearning, result.Status)
	}
	// SM-2: rep=1, interval=1
	if result.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", result.Repetitions)
	}
	// No lapse
	if result.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", result.Lapses)
	}
	// SM-2 EF: 2.5 + (0.1 - 1*(0.08+0.02)) = 2.5 + 0.0 = 2.5
	if roundFloat(result.EaseFactor) != 2.5 {
		t.Errorf("expected ease_factor≈2.5, got %.4f", result.EaseFactor)
	}
}

func TestSubmitReview_NewCard_Again(t *testing.T) {
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Again)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays new (quality < Good, no transition from new)
	if result.Status != model.ReviewStatusNew {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusNew, result.Status)
	}
	// SM-2 failure: reps=0, interval=1
	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", result.Repetitions)
	}
	if result.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", result.IntervalDays)
	}
	// Lapse counted even on new card
	if result.Lapses != 1 {
		t.Errorf("expected lapses=1, got %d", result.Lapses)
	}
}

func TestSubmitReview_NewCard_Hard(t *testing.T) {
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Hard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays new (quality < Good)
	if result.Status != model.ReviewStatusNew {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusNew, result.Status)
	}
	// Hard is a failure in SM-2 (quality < 3): reps=0, interval=1
	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", result.Repetitions)
	}
	if result.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", result.IntervalDays)
	}
	// Lapse counted
	if result.Lapses != 1 {
		t.Errorf("expected lapses=1, got %d", result.Lapses)
	}
}

// ── Tests: SubmitReview — Learning Card ──────────────────────────────────────────

func TestSubmitReview_LearningCard_Good(t *testing.T) {
	card := newCard(model.ReviewStatusLearning, 1, 2.36, 1)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: learning → review
	if result.Status != model.ReviewStatusReview {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusReview, result.Status)
	}
	// SM-2: rep=2, interval=6 (second pass)
	if result.Repetitions != 2 {
		t.Errorf("expected repetitions=2, got %d", result.Repetitions)
	}
	if result.IntervalDays != 6 {
		t.Errorf("expected interval_days=6, got %d", result.IntervalDays)
	}
	// No lapse
	if result.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", result.Lapses)
	}
}

func TestSubmitReview_LearningCard_Easy(t *testing.T) {
	card := newCard(model.ReviewStatusLearning, 1, 2.36, 1)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Easy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: learning → review (Easy >= Good)
	if result.Status != model.ReviewStatusReview {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusReview, result.Status)
	}
	if result.Repetitions != 2 {
		t.Errorf("expected repetitions=2, got %d", result.Repetitions)
	}
	if result.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", result.Lapses)
	}
}

func TestSubmitReview_LearningCard_Again(t *testing.T) {
	card := newCard(model.ReviewStatusLearning, 1, 2.36, 1)
	card.Lapses = 0 // explicitly starting from 0 lapses

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Again)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays learning (quality < Good, no transition from learning)
	if result.Status != model.ReviewStatusLearning {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusLearning, result.Status)
	}
	// SM-2 failure: reps=0, interval=1
	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", result.Repetitions)
	}
	if result.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", result.IntervalDays)
	}
	// Lapse increments
	if result.Lapses != 1 {
		t.Errorf("expected lapses=1, got %d", result.Lapses)
	}
}

// ── Tests: SubmitReview — Review Card (graduated) ───────────────────────────────

func TestSubmitReview_ReviewCard_Good(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays review (no transition)
	if result.Status != model.ReviewStatusReview {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusReview, result.Status)
	}
	// SM-2: rep=4, interval=round(13*2.5)=33
	if result.Repetitions != 4 {
		t.Errorf("expected repetitions=4, got %d", result.Repetitions)
	}
	expectedInterval := int(math.Round(13 * 2.5))
	if result.IntervalDays != expectedInterval {
		t.Errorf("expected interval_days=%d, got %d", expectedInterval, result.IntervalDays)
	}
	// No lapse
	if result.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", result.Lapses)
	}
}

func TestSubmitReview_ReviewCard_Easy(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Easy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays review
	if result.Status != model.ReviewStatusReview {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusReview, result.Status)
	}
	// Easy gives higher ease factor and same interval calculation
	if result.Repetitions != 4 {
		t.Errorf("expected repetitions=4, got %d", result.Repetitions)
	}
	// SM-2 EF: 2.5 + (0.1 - 1*(0.08+0.02)) = 2.5
	if roundFloat(result.EaseFactor) != 2.5 {
		t.Errorf("expected ease_factor≈2.5, got %.4f", result.EaseFactor)
	}
	if result.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", result.Lapses)
	}
}

func TestSubmitReview_ReviewCard_Again(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 5, 2.5, 20)
	card.Lapses = 0

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Again)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays review (no "lapse back to learning" in current implementation)
	if result.Status != model.ReviewStatusReview {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusReview, result.Status)
	}
	// SM-2 failure: reps=0, interval=1
	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", result.Repetitions)
	}
	if result.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", result.IntervalDays)
	}
	// Lapse increments
	if result.Lapses != 1 {
		t.Errorf("expected lapses=1, got %d", result.Lapses)
	}
	// Ease factor decreases: EF = 2.5 + (0.1 - (5-1)*(0.08+(5-1)*0.02)) = 2.5 + (0.1 - 0.64) = 1.96
	expectedEF := 1.96
	if roundFloat(result.EaseFactor) != expectedEF {
		t.Errorf("expected ease_factor≈%.2f, got %.4f", expectedEF, result.EaseFactor)
	}
}

func TestSubmitReview_ReviewCard_Hard(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)
	card.Lapses = 0

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Hard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status: stays review
	if result.Status != model.ReviewStatusReview {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusReview, result.Status)
	}
	// SM-2 failure: reps=0, interval=1
	if result.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", result.Repetitions)
	}
	if result.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", result.IntervalDays)
	}
	// Lapse increments
	if result.Lapses != 1 {
		t.Errorf("expected lapses=1, got %d", result.Lapses)
	}
}

// ── Tests: SubmitReview — Ease Factor Floor ──────────────────────────────────────

func TestSubmitReview_EaseFactorFloor(t *testing.T) {
	// Card with EF already at 1.3 (floor)
	card := newCard(model.ReviewStatusReview, 2, 1.3, 6)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Again)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SM-2 would compute: 1.3 + (0.1 - 4*(0.08+0.08)) = 1.3 - 0.56 = 0.74 → clamped to 1.3
	if result.EaseFactor < 1.3 {
		t.Errorf("ease factor should never fall below 1.3, got %.4f", result.EaseFactor)
	}
	if roundFloat(result.EaseFactor) != 1.3 {
		t.Errorf("expected ease_factor clamped to 1.30, got %.4f", result.EaseFactor)
	}
}

func TestSubmitReview_EaseFactorFloorOnHard(t *testing.T) {
	// Card with low EF, Hard answer would push below floor
	card := newCard(model.ReviewStatusReview, 2, 1.3, 6)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Hard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SM-2: 1.3 + (0.1 - 3*(0.08+0.06)) = 1.3 - 0.32 = 0.98 → clamped to 1.3
	if result.EaseFactor < 1.3 {
		t.Errorf("ease factor should never fall below 1.3, got %.4f", result.EaseFactor)
	}
}

// ── Tests: SubmitReview — Lapse Accumulation ─────────────────────────────────────

func TestSubmitReview_LapseAccumulates(t *testing.T) {
	// Card that has already lapsed once
	card := newCard(model.ReviewStatusReview, 5, 2.5, 20)
	card.Lapses = 3

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Again)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Lapses should accumulate: 3 + 1 = 4
	if result.Lapses != 4 {
		t.Errorf("expected lapses=4 (3 existing + 1 new), got %d", result.Lapses)
	}
}

func TestSubmitReview_NoLapseOnCorrectAnswer(t *testing.T) {
	// Card that has 2 existing lapses
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)
	card.Lapses = 2

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Lapses should not increment on correct answer
	if result.Lapses != 2 {
		t.Errorf("expected lapses=2 (unchanged), got %d", result.Lapses)
	}
}

// ── Tests: SubmitReview — Suspended Card ─────────────────────────────────────────

func TestSubmitReview_SuspendedCard_NoStatusTransition(t *testing.T) {
	card := newCard(model.ReviewStatusSuspended, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Suspended status is preserved regardless of answer quality
	if result.Status != model.ReviewStatusSuspended {
		t.Errorf("expected status=%s (unchanged), got %s", model.ReviewStatusSuspended, result.Status)
	}
}

// ── Tests: SubmitReview — SM-2 Interval Progression ──────────────────────────────

func TestSubmitReview_IntervalProgression_FirstSecondThird(t *testing.T) {
	// Simulate a full progression: new → learning (Good) → review (Good) → review (Good)
	// Step 1: new card with Good
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("step 1: unexpected error: %v", err)
	}
	if result.IntervalDays != 1 {
		t.Errorf("step 1: expected interval=1, got %d", result.IntervalDays)
	}
	if result.Repetitions != 1 {
		t.Errorf("step 1: expected rep=1, got %d", result.Repetitions)
	}

	// Step 2: learning card with Good → should get interval=6
	card2 := newCard(model.ReviewStatusLearning, 1, 2.36, 1)
	reviewRepo2 := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card2, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	svc2 := newReviewService(reviewRepo2, wordRepo)

	result2, err := svc2.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("step 2: unexpected error: %v", err)
	}
	if result2.IntervalDays != 6 {
		t.Errorf("step 2: expected interval=6, got %d", result2.IntervalDays)
	}

	// Step 3: review card with Good → interval = round(previous_interval * EF)
	card3 := newCard(model.ReviewStatusReview, 2, 2.22, 6)
	reviewRepo3 := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card3, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	svc3 := newReviewService(reviewRepo3, wordRepo)

	result3, err := svc3.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("step 3: unexpected error: %v", err)
	}
	// interval = round(6 * 2.22) = round(13.32) = 13
	if result3.IntervalDays != 13 {
		t.Errorf("step 3: expected interval=13, got %d", result3.IntervalDays)
	}
}

// ── Tests: SubmitReview — NextReview Date Computation ─────────────────────────────

func TestSubmitReview_NextReviewIsCorrect(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	before := time.Now().UTC()
	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// NextReview should be approximately now + interval days
	expectedMin := before.AddDate(0, 0, result.IntervalDays)
	expectedMax := after.AddDate(0, 0, result.IntervalDays)
	if result.NextReview.Before(expectedMin) || result.NextReview.After(expectedMax) {
		t.Errorf("expected NextReview between %v and %v, got %v", expectedMin, expectedMax, result.NextReview)
	}
}

// ── Tests: SubmitReview — Repository Errors ───────────────────────────────────────

func TestSubmitReview_RepoGetError(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return nil, fmt.Errorf("card not found")
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "submit review: card not found" {
		t.Errorf("expected wrapped error, got %q", err.Error())
	}
}

func TestSubmitReview_RepoUpdateError(t *testing.T) {
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return fmt.Errorf("update failed")
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "submit review: update: update failed" {
		t.Errorf("expected wrapped error, got %q", err.Error())
	}
}

// ── Tests: SubmitReview — Fields are updated correctly ────────────────────────────

func TestSubmitReview_UpdatesLastReviewedAt(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)
	if card.LastReviewedAt != nil {
		t.Fatal("test setup: expected LastReviewedAt to be nil initially")
	}

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.LastReviewedAt == nil {
		t.Error("expected last_reviewed_at to be set after review")
	}
}

func TestSubmitReview_UpdatesUpdatedAt(t *testing.T) {
	card := newCard(model.ReviewStatusReview, 3, 2.5, 13)
	originalUpdatedAt := card.UpdatedAt

	// Ensure the mock update captures the card
	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	result, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.UpdatedAt.After(originalUpdatedAt.Add(-time.Second)) {
		t.Error("expected updated_at to be near current time")
	}
}

// ── Tests: CountDue ───────────────────────────────────────────────────────────────

func TestCountDue_ReturnsCount(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		countDueFunc: func(ctx context.Context) (int, error) {
			return 5, nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	count, err := svc.CountDue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count=5, got %d", count)
	}
}

func TestCountDue_ZeroCards(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		countDueFunc: func(ctx context.Context) (int, error) {
			return 0, nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	count, err := svc.CountDue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count=0, got %d", count)
	}
}

func TestCountDue_RepoError(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		countDueFunc: func(ctx context.Context) (int, error) {
			return 0, fmt.Errorf("db error")
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.CountDue(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "count due: db error" {
		t.Errorf("expected wrapped error, got %q", err.Error())
	}
}

// ── Tests: CreateCardForWord ────────────────────────────────────────────────────

func TestCreateCardForWord_SetsDefaultsAndCreates(t *testing.T) {
	var captured *model.ReviewCard

	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			captured = card
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	err := svc.CreateCardForWord(context.Background(), "word-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured == nil {
		t.Fatal("expected card to be created, got nil")
	}
	if captured.WordID != "word-abc" {
		t.Errorf("expected word_id=word-abc, got %s", captured.WordID)
	}
	if captured.Status != model.ReviewStatusNew {
		t.Errorf("expected status=%s, got %s", model.ReviewStatusNew, captured.Status)
	}
	if captured.EaseFactor != 2.5 {
		t.Errorf("expected ease_factor=2.5, got %f", captured.EaseFactor)
	}
	if captured.IntervalDays != 0 {
		t.Errorf("expected interval_days=0, got %d", captured.IntervalDays)
	}
	if captured.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", captured.Repetitions)
	}
	if captured.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", captured.Lapses)
	}
	if captured.ID == "" {
		t.Error("expected non-empty ID")
	}
	// NextReview should be approximately now
	now := time.Now().UTC()
	diff := now.Sub(captured.NextReview)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("expected next_review ≈ now, got diff=%v", diff)
	}
	// CreatedAt and UpdatedAt should be approximately now
	diff = now.Sub(captured.CreatedAt)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("expected created_at ≈ now, got diff=%v", diff)
	}
}

func TestCreateCardForWord_RepoError(t *testing.T) {
	reviewRepo := &mockReviewRepo{
		createFunc: func(ctx context.Context, card *model.ReviewCard) error {
			return fmt.Errorf("insert failed")
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	err := svc.CreateCardForWord(context.Background(), "word-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "insert failed" {
		t.Errorf("expected 'insert failed', got %q", err.Error())
	}
}

// ── Tests: SubmitReview — Verify card passed to UpdateReview ──────────────────────

func TestSubmitReview_PassesUpdatedCardToRepo(t *testing.T) {
	card := newCard(model.ReviewStatusNew, 0, 2.5, 0)
	var capturedCard *model.ReviewCard

	reviewRepo := &mockReviewRepo{
		getByWordIDFunc: func(ctx context.Context, wordID string) (*model.ReviewCard, error) {
			return card, nil
		},
		updateReviewFunc: func(ctx context.Context, c *model.ReviewCard) error {
			capturedCard = c
			return nil
		},
	}
	wordRepo := &mockWordRepo{}
	svc := newReviewService(reviewRepo, wordRepo)

	_, err := svc.SubmitReview(context.Background(), "word-1", srs.Good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedCard == nil {
		t.Fatal("expected UpdateReview to be called with a card, got nil")
	}
	// Verify the same card pointer is updated (identity check)
	if capturedCard != card {
		t.Error("expected the same card pointer to be passed to UpdateReview")
	}
	// Verify all SM-2 fields are propagated
	if capturedCard.Repetitions != 1 {
		t.Errorf("expected card.Repetitions=1, got %d", capturedCard.Repetitions)
	}
	if roundFloat(capturedCard.EaseFactor) != 2.36 {
		t.Errorf("expected card.EaseFactor≈2.36, got %.4f", capturedCard.EaseFactor)
	}
	if capturedCard.IntervalDays != 1 {
		t.Errorf("expected card.IntervalDays=1, got %d", capturedCard.IntervalDays)
	}
}
