package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SavedWord represents a word saved for study.
type SavedWord struct {
	ID          string    `json:"id"`
	DocumentID  string    `json:"document_id"`
	Word        string    `json:"word"`
	Translation string    `json:"translation"`
	SourceLang  string    `json:"source_lang"`
	TargetLang  string    `json:"target_lang"`
	CreatedAt   time.Time `json:"created_at"`
}

// WordRepository defines persistence operations for saved words.
type WordRepository interface {
	Save(ctx context.Context, word *SavedWord) error
	ListByDocument(ctx context.Context, documentID string) ([]*SavedWord, error)
	ListAll(ctx context.Context) ([]*SavedWord, error)
	Delete(ctx context.Context, id string) error
}

type wordRepo struct {
	db *sql.DB
}

// NewWordRepository creates a new WordRepository backed by SQLite.
func NewWordRepository(db *sql.DB) WordRepository {
	return &wordRepo{db: db}
}

func (r *wordRepo) Save(ctx context.Context, word *SavedWord) error {
	query := `INSERT INTO saved_words (id, document_id, word, translation, source_lang, target_lang, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, word.ID, word.DocumentID, word.Word, word.Translation, word.SourceLang, word.TargetLang, word.CreatedAt)
	if err != nil {
		return fmt.Errorf("save word: %w", err)
	}
	return nil
}

func (r *wordRepo) ListByDocument(ctx context.Context, documentID string) ([]*SavedWord, error) {
	query := `SELECT id, document_id, word, translation, source_lang, target_lang, created_at FROM saved_words WHERE document_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("list words: %w", err)
	}
	defer rows.Close()

	var words []*SavedWord
	for rows.Next() {
		w := &SavedWord{}
		var createdRaw any
		if err := rows.Scan(&w.ID, &w.DocumentID, &w.Word, &w.Translation, &w.SourceLang, &w.TargetLang, &createdRaw); err != nil {
			return nil, fmt.Errorf("scan word: %w", err)
		}
		createdAt, err := parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse time: %w", err)
		}
		w.CreatedAt = createdAt
		words = append(words, w)
	}
	return words, rows.Err()
}

func (r *wordRepo) ListAll(ctx context.Context) ([]*SavedWord, error) {
	query := `SELECT id, document_id, word, translation, source_lang, target_lang, created_at FROM saved_words ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all words: %w", err)
	}
	defer rows.Close()

	var words []*SavedWord
	for rows.Next() {
		w := &SavedWord{}
		var createdRaw any
		if err := rows.Scan(&w.ID, &w.DocumentID, &w.Word, &w.Translation, &w.SourceLang, &w.TargetLang, &createdRaw); err != nil {
			return nil, fmt.Errorf("scan word: %w", err)
		}
		createdAt, err := parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse time: %w", err)
		}
		w.CreatedAt = createdAt
		words = append(words, w)
	}
	return words, rows.Err()
}

func (r *wordRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM saved_words WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete word: %w", err)
	}
	return nil
}
