package repository

import "context"

type WordRepoMock struct {
	words []*SavedWord
}

func NewWordRepoMock() *WordRepoMock {
	return &WordRepoMock{}
}

func (m *WordRepoMock) Seed(words []*SavedWord) {
	m.words = words
}

func (m *WordRepoMock) Save(ctx context.Context, word *SavedWord) error {
	m.words = append(m.words, word)
	return nil
}

func (m *WordRepoMock) ListByDocument(ctx context.Context, documentID string) ([]*SavedWord, error) {
	return m.words, nil
}

func (m *WordRepoMock) ListAll(ctx context.Context) ([]*SavedWord, error) {
	return m.words, nil
}

func (m *WordRepoMock) Delete(ctx context.Context, id string) error {
	return nil
}
