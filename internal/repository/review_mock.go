package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/croko/language-app/internal/model"
)

// ReviewRepoMock implements ReviewRepository with an in-memory store for testing.
type ReviewRepoMock struct {
	mu    sync.Mutex
	cards map[string]*model.ReviewCard // keyed by word_id
}

// NewReviewRepoMock creates a new ReviewRepoMock.
func NewReviewRepoMock() *ReviewRepoMock {
	return &ReviewRepoMock{
		cards: make(map[string]*model.ReviewCard),
	}
}

// Seed adds cards directly (for test setup).
func (m *ReviewRepoMock) Seed(cards []*model.ReviewCard) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range cards {
		m.cards[c.WordID] = c
	}
}

func (m *ReviewRepoMock) Create(_ context.Context, card *model.ReviewCard) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.cards[card.WordID]; ok {
		return errors.New("duplicate word_id")
	}
	m.cards[card.WordID] = card
	return nil
}

func (m *ReviewRepoMock) GetByWordID(_ context.Context, wordID string) (*model.ReviewCard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	card, ok := m.cards[wordID]
	if !ok {
		return nil, errors.New("review card not found")
	}
	return card, nil
}

func (m *ReviewRepoMock) GetDueWords(_ context.Context) ([]*model.ReviewCard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return all cards — tests can seed the right state
	var result []*model.ReviewCard
	for _, card := range m.cards {
		result = append(result, card)
	}
	return result, nil
}

func (m *ReviewRepoMock) UpdateReview(_ context.Context, card *model.ReviewCard) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.cards[card.WordID]; !ok {
		return errors.New("review card not found")
	}
	m.cards[card.WordID] = card
	return nil
}

func (m *ReviewRepoMock) CountDue(_ context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return number of cards (tests seed the expected count)
	return len(m.cards), nil
}

func (m *ReviewRepoMock) DeleteByDocumentID(_ context.Context, _ string) error {
	return nil
}
