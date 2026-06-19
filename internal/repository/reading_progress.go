package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/google/uuid"
)

type ReadingProgressRepository interface {
	Upsert(ctx context.Context, documentID string, chapterIndex int, percentage float64) (*model.ReadingProgress, error)
	GetByDocumentID(ctx context.Context, documentID string) (*model.ReadingProgress, error)
	DeleteByDocumentID(ctx context.Context, documentID string) error
}

type readingProgressRepo struct {
	db *sql.DB
}

func NewReadingProgressRepository(db *sql.DB) ReadingProgressRepository {
	return &readingProgressRepo{db: db}
}

func (r *readingProgressRepo) Upsert(ctx context.Context, documentID string, chapterIndex int, percentage float64) (*model.ReadingProgress, error) {
	now := time.Now().UTC()
	id := uuid.New().String()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO reading_progress (id, document_id, chapter_index, percentage, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(document_id) DO UPDATE SET
			chapter_index = excluded.chapter_index,
			percentage = excluded.percentage,
			updated_at = excluded.updated_at
	`, id, documentID, chapterIndex, percentage, now)
	if err != nil {
		return nil, fmt.Errorf("upsert reading progress: %w", err)
	}

	return r.GetByDocumentID(ctx, documentID)
}

func (r *readingProgressRepo) GetByDocumentID(ctx context.Context, documentID string) (*model.ReadingProgress, error) {
	query := `
		SELECT id, document_id, chapter_index, percentage, updated_at
		FROM reading_progress
		WHERE document_id = ?
	`
	p := &model.ReadingProgress{}
	var updatedRaw any
	err := r.db.QueryRowContext(ctx, query, documentID).Scan(
		&p.ID,
		&p.DocumentID,
		&p.ChapterIndex,
		&p.Percentage,
		&updatedRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reading progress not found: %w", err)
		}
		return nil, fmt.Errorf("get reading progress: %w", err)
	}

	updatedAt, err := parseDBTime(updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	p.UpdatedAt = updatedAt

	return p, nil
}

func (r *readingProgressRepo) DeleteByDocumentID(ctx context.Context, documentID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM reading_progress WHERE document_id = ?", documentID)
	if err != nil {
		return fmt.Errorf("delete reading progress: %w", err)
	}
	return nil
}
