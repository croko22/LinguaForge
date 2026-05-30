package service

import (
	"context"
	"time"

	"github.com/croko/language-app/internal/repository"
	"github.com/google/uuid"
)

type WordService struct {
	repo repository.WordRepository
}

func NewWordService(repo repository.WordRepository) *WordService {
	return &WordService{repo: repo}
}

func (s *WordService) SaveWord(ctx context.Context, documentID, word, translation, sourceLang, targetLang string) (*repository.SavedWord, error) {
	sw := &repository.SavedWord{
		ID:          uuid.New().String(),
		DocumentID:  documentID,
		Word:        word,
		Translation: translation,
		SourceLang:  sourceLang,
		TargetLang:  targetLang,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.Save(ctx, sw); err != nil {
		return nil, err
	}
	return sw, nil
}

func (s *WordService) ListWords(ctx context.Context) ([]*repository.SavedWord, error) {
	return s.repo.ListAll(ctx)
}

func (s *WordService) DeleteWord(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
