package service

import (
	"context"
	"fmt"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/translator"
	"github.com/google/uuid"
)

type WordService struct {
	repo       repository.WordRepository
	reviewRepo repository.ReviewRepository
	translator translator.Translator
}

func NewWordService(repo repository.WordRepository, reviewRepo repository.ReviewRepository, t translator.Translator) *WordService {
	return &WordService{
		repo:       repo,
		reviewRepo: reviewRepo,
		translator: t,
	}
}

func (s *WordService) SaveWord(ctx context.Context, documentID, word, translation, sourceLang, targetLang string) (*model.SavedWord, error) {
	// If no translation provided, call the translator
	if translation == "" {
		resp, err := s.translator.Translate(ctx, translator.TranslateRequest{
			Word:       word,
			SourceLang: sourceLang,
			TargetLang: targetLang,
		})
		if err == nil {
			translation = resp.Translation
		}
		// If translation fails, continue with empty — don't block the save
	}

	now := time.Now().UTC()
	sw := &model.SavedWord{
		ID:          uuid.New().String(),
		DocumentID:  documentID,
		Word:        word,
		Translation: translation,
		SourceLang:  sourceLang,
		TargetLang:  targetLang,
		CreatedAt:   now,
	}
	if err := s.repo.Save(ctx, sw); err != nil {
		return nil, fmt.Errorf("save word: %w", err)
	}

	// Auto-create a review card for the new word
	card := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       sw.ID,
		Status:       model.ReviewStatusNew,
		EaseFactor:   2.5,
		IntervalDays: 0,
		Repetitions:  0,
		Lapses:       0,
		NextReview:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.reviewRepo.Create(ctx, card); err != nil {
		// Don't fail the word save if review card creation fails
		return sw, nil
	}

	return sw, nil
}

func (s *WordService) ListWords(ctx context.Context) ([]*model.SavedWord, error) {
	return s.repo.ListAll(ctx)
}

func (s *WordService) DeleteWord(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
