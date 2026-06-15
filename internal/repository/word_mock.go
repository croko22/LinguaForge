package repository

import (
	"context"

	"github.com/croko/language-app/internal/model"
)

type WordRepoMock struct {
	words []*model.SavedWord
}

func NewWordRepoMock() *WordRepoMock {
	return &WordRepoMock{}
}

func (m *WordRepoMock) Seed(words []*model.SavedWord) {
	m.words = words
}

func (m *WordRepoMock) Save(ctx context.Context, word *model.SavedWord) error {
	m.words = append(m.words, word)
	return nil
}

func (m *WordRepoMock) ListByDocument(ctx context.Context, documentID string) ([]*model.SavedWord, error) {
	return m.words, nil
}

func (m *WordRepoMock) ListAll(ctx context.Context) ([]*model.SavedWord, error) {
	return m.words, nil
}

func (m *WordRepoMock) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *WordRepoMock) DeleteByDocumentID(ctx context.Context, documentID string) error {
	return nil
}
