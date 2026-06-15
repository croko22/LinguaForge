package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/croko/language-app/internal/model"
)

// sqliteTimeFmt is the format used by SQLite's datetime() function.
const sqliteTimeFmt = "2006-01-02 15:04:05"

// ReviewRepository defines persistence operations for review cards.
type ReviewRepository interface {
	// Create inserts a new review card for a saved word.
	Create(ctx context.Context, card *model.ReviewCard) error
	// GetByWordID retrieves the review card for a specific word.
	GetByWordID(ctx context.Context, wordID string) (*model.ReviewCard, error)
	// GetDueWords returns all review cards due for review (next_review <= now).
	GetDueWords(ctx context.Context) ([]*model.ReviewCard, error)
	// UpdateReview updates a review card after a review submission.
	UpdateReview(ctx context.Context, card *model.ReviewCard) error
	// CountDue returns the number of review cards due for review.
	CountDue(ctx context.Context) (int, error)
	// DeleteByDocumentID deletes all review cards for words belonging to the given document.
	DeleteByDocumentID(ctx context.Context, documentID string) error
}

type reviewRepo struct {
	db *sql.DB
}

// NewReviewRepository creates a new ReviewRepository backed by SQLite.
func NewReviewRepository(db *sql.DB) ReviewRepository {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(ctx context.Context, card *model.ReviewCard) error {
	query := `
		INSERT INTO word_reviews (id, word_id, status, ease_factor, interval_days, repetitions, lapses, next_review, last_reviewed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		card.ID,
		card.WordID,
		card.Status,
		card.EaseFactor,
		card.IntervalDays,
		card.Repetitions,
		card.Lapses,
		card.NextReview.Format(sqliteTimeFmt),
		sqliteNullableTime(card.LastReviewedAt),
		card.CreatedAt.Format(sqliteTimeFmt),
		card.UpdatedAt.Format(sqliteTimeFmt),
	)
	if err != nil {
		return fmt.Errorf("create review card: %w", err)
	}
	return nil
}

func (r *reviewRepo) GetByWordID(ctx context.Context, wordID string) (*model.ReviewCard, error) {
	query := `
		SELECT id, word_id, status, ease_factor, interval_days, repetitions, lapses,
		       next_review, last_reviewed_at, created_at, updated_at
		FROM word_reviews
		WHERE word_id = ?
	`
	card := &model.ReviewCard{}
	var nextRaw any
	var lastRaw any
	var createdRaw any
	var updatedRaw any

	err := r.db.QueryRowContext(ctx, query, wordID).Scan(
		&card.ID,
		&card.WordID,
		&card.Status,
		&card.EaseFactor,
		&card.IntervalDays,
		&card.Repetitions,
		&card.Lapses,
		&nextRaw,
		&lastRaw,
		&createdRaw,
		&updatedRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("review card not found for word %s: %w", wordID, sql.ErrNoRows)
		}
		return nil, fmt.Errorf("get review card by word id: %w", err)
	}

	card.NextReview, err = parseDBTime(nextRaw)
	if err != nil {
		return nil, fmt.Errorf("parse next_review: %w", err)
	}
	card.LastReviewedAt, err = parseNullableDBTime(lastRaw)
	if err != nil {
		return nil, fmt.Errorf("parse last_reviewed_at: %w", err)
	}
	card.CreatedAt, err = parseDBTime(createdRaw)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	card.UpdatedAt, err = parseDBTime(updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return card, nil
}

func (r *reviewRepo) GetDueWords(ctx context.Context) ([]*model.ReviewCard, error) {
	query := `
		SELECT id, word_id, status, ease_factor, interval_days, repetitions, lapses,
		       next_review, last_reviewed_at, created_at, updated_at
		FROM word_reviews
		WHERE next_review <= datetime('now')
		ORDER BY next_review ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get due words: %w", err)
	}
	defer rows.Close()

	var cards []*model.ReviewCard
	for rows.Next() {
		card := &model.ReviewCard{}
		var nextRaw any
		var lastRaw any
		var createdRaw any
		var updatedRaw any

		if err := rows.Scan(
			&card.ID,
			&card.WordID,
			&card.Status,
			&card.EaseFactor,
			&card.IntervalDays,
			&card.Repetitions,
			&card.Lapses,
			&nextRaw,
			&lastRaw,
			&createdRaw,
			&updatedRaw,
		); err != nil {
			return nil, fmt.Errorf("scan due word: %w", err)
		}

		card.NextReview, err = parseDBTime(nextRaw)
		if err != nil {
			return nil, fmt.Errorf("parse next_review: %w", err)
		}
		card.LastReviewedAt, err = parseNullableDBTime(lastRaw)
		if err != nil {
			return nil, fmt.Errorf("parse last_reviewed_at: %w", err)
		}
		card.CreatedAt, err = parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		card.UpdatedAt, err = parseDBTime(updatedRaw)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at: %w", err)
		}

		cards = append(cards, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get due words rows: %w", err)
	}
	return cards, nil
}

func (r *reviewRepo) UpdateReview(ctx context.Context, card *model.ReviewCard) error {
	query := `
		UPDATE word_reviews
		SET status = ?, ease_factor = ?, interval_days = ?, repetitions = ?, lapses = ?,
		    next_review = ?, last_reviewed_at = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		card.Status,
		card.EaseFactor,
		card.IntervalDays,
		card.Repetitions,
		card.Lapses,
		card.NextReview.Format(sqliteTimeFmt),
		sqliteNullableTime(card.LastReviewedAt),
		card.UpdatedAt.Format(sqliteTimeFmt),
		card.ID,
	)
	if err != nil {
		return fmt.Errorf("update review card: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update review card rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("review card not found: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *reviewRepo) CountDue(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM word_reviews WHERE next_review <= datetime('now')`
	var count int
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count due words: %w", err)
	}
	return count, nil
}

func (r *reviewRepo) DeleteByDocumentID(ctx context.Context, documentID string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM word_reviews WHERE word_id IN (SELECT id FROM saved_words WHERE document_id = ?)",
		documentID,
	)
	if err != nil {
		return fmt.Errorf("delete word reviews by document: %w", err)
	}
	return nil
}

// sqliteNullableTime formats a *time.Time pointer as SQLite text format,
// returning nil if the pointer is nil (stores as SQL NULL).
func sqliteNullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(sqliteTimeFmt)
}

// parseNullableDBTime parses a nullable database time value into a *time.Time pointer.
func parseNullableDBTime(raw any) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	t, err := parseDBTime(raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
