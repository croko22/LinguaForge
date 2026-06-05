package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/srs"
	"github.com/google/uuid"
)

// ReviewService provides business logic for spaced repetition review operations.
type ReviewService struct {
	reviewRepo repository.ReviewRepository
	wordRepo   repository.WordRepository
}

// NewReviewService creates a new ReviewService.
func NewReviewService(reviewRepo repository.ReviewRepository, wordRepo repository.WordRepository) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		wordRepo:   wordRepo,
	}
}

// DueWordResponse is the DTO for a due review card with word data.
type DueWordResponse struct {
	ID           string    `json:"id"`            // word ID
	Word         string    `json:"word"`          // from SavedWord
	Translation  string    `json:"translation"`   // from SavedWord
	SourceLang   string    `json:"source_lang"`   // from SavedWord
	TargetLang   string    `json:"target_lang"`   // from SavedWord
	DocumentID   string    `json:"document_id"`   // from SavedWord
	Status       string    `json:"status"`        // from ReviewCard
	NextReview   time.Time `json:"next_review"`   // from ReviewCard
	EaseFactor   float64   `json:"ease_factor"`   // from ReviewCard
	IntervalDays int       `json:"interval_days"` // from ReviewCard
	Repetitions  int       `json:"repetitions"`   // from ReviewCard
	Lapses       int       `json:"lapses"`        // from ReviewCard
}

// GetDue returns review cards that are due for review, joined with word data.
func (s *ReviewService) GetDue(ctx context.Context) ([]*DueWordResponse, error) {
	cards, err := s.reviewRepo.GetDueWords(ctx)
	if err != nil {
		return nil, fmt.Errorf("get due: %w", err)
	}

	if len(cards) == 0 {
		return []*DueWordResponse{}, nil
	}

	// Build word lookup
	words, err := s.wordRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get due: list words: %w", err)
	}
	wordMap := make(map[string]*model.SavedWord, len(words))
	for _, w := range words {
		wordMap[w.ID] = w
	}

	result := make([]*DueWordResponse, 0, len(cards))
	for _, card := range cards {
		w, ok := wordMap[card.WordID]
		if !ok {
			continue // skip cards without word data (shouldn't happen)
		}
		result = append(result, &DueWordResponse{
			ID:           card.WordID,
			Word:         w.Word,
			Translation:  w.Translation,
			SourceLang:   w.SourceLang,
			TargetLang:   w.TargetLang,
			DocumentID:   w.DocumentID,
			Status:       card.Status,
			NextReview:   card.NextReview,
			EaseFactor:   card.EaseFactor,
			IntervalDays: card.IntervalDays,
			Repetitions:  card.Repetitions,
			Lapses:       card.Lapses,
		})
	}

	return result, nil
}

// SubmitReview processes a review: runs SM-2 algorithm and updates the card.
func (s *ReviewService) SubmitReview(ctx context.Context, wordID string, quality srs.Quality) (*model.ReviewCard, error) {
	if quality < srs.Again || quality > srs.Easy {
		return nil, fmt.Errorf("invalid quality: %d: %w", quality, ErrInvalidQuality)
	}

	card, err := s.reviewRepo.GetByWordID(ctx, wordID)
	if err != nil {
		return nil, fmt.Errorf("submit review: %w", err)
	}

	now := time.Now().UTC()

	input := srs.CardInput{
		Repetitions: card.Repetitions,
		EaseFactor:  card.EaseFactor,
		Interval:    card.IntervalDays,
	}

	output := srs.ComputeNextReview(input, quality, now)

	// Update card fields from SM-2 output
	card.Repetitions = output.Repetitions
	card.EaseFactor = output.EaseFactor
	card.IntervalDays = output.Interval
	card.NextReview = output.NextReview
	card.UpdatedAt = now
	card.LastReviewedAt = &now

	// Handle failure: increment lapses
	if quality < srs.Good {
		card.Lapses++
	}

	// Status transitions
	switch card.Status {
	case model.ReviewStatusNew:
		if quality >= srs.Good {
			card.Status = model.ReviewStatusLearning
		}
	case model.ReviewStatusLearning:
		if quality >= srs.Good {
			card.Status = model.ReviewStatusReview
		}
	case model.ReviewStatusReview, model.ReviewStatusSuspended:
		// Keep current status
	}

	if err := s.reviewRepo.UpdateReview(ctx, card); err != nil {
		return nil, fmt.Errorf("submit review: update: %w", err)
	}

	return card, nil
}

// CountDue returns the number of cards due for review.
func (s *ReviewService) CountDue(ctx context.Context) (int, error) {
	count, err := s.reviewRepo.CountDue(ctx)
	if err != nil {
		return 0, fmt.Errorf("count due: %w", err)
	}
	return count, nil
}

// CreateCardForWord creates a review card for an existing saved word.
func (s *ReviewService) CreateCardForWord(ctx context.Context, wordID string) error {
	now := time.Now().UTC()
	card := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       wordID,
		Status:       model.ReviewStatusNew,
		EaseFactor:   2.5,
		IntervalDays: 0,
		Repetitions:  0,
		Lapses:       0,
		NextReview:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.reviewRepo.Create(ctx, card)
}

// ErrInvalidQuality is returned when a quality value is outside the valid range.
var ErrInvalidQuality = errors.New("quality must be between 1 and 4")
