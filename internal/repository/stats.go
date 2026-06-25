package repository

import (
	"context"
	"fmt"
)

// ReviewActivity represents a single day's review count for the heatmap.
type ReviewActivity struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int    `json:"count"`
}

// GetReviewActivity returns review counts per day for the last N days.
func (r *reviewRepo) GetReviewActivity(ctx context.Context, days int) ([]ReviewActivity, error) {
	query := `
		SELECT date(last_reviewed_at) AS day, COUNT(*) AS count
		FROM word_reviews
		WHERE last_reviewed_at >= datetime('now', ?)
		GROUP BY day
		ORDER BY day ASC
	`
	offset := fmt.Sprintf("-%d days", days)

	rows, err := r.db.QueryContext(ctx, query, offset)
	if err != nil {
		return nil, fmt.Errorf("get review activity: %w", err)
	}
	defer rows.Close()

	var activity []ReviewActivity
	for rows.Next() {
		var a ReviewActivity
		if err := rows.Scan(&a.Date, &a.Count); err != nil {
			return nil, fmt.Errorf("scan review activity: %w", err)
		}
		activity = append(activity, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("review activity rows: %w", err)
	}

	if activity == nil {
		activity = []ReviewActivity{}
	}

	return activity, nil
}

type StatsSnapshot struct {
	TotalDocuments int              `json:"total_documents"`
	TotalWords     int              `json:"total_words"`
	TotalChapters  int              `json:"total_chapters"`
	LanguageCounts []LanguageCount  `json:"language_counts"`
	ReviewActivity []ReviewActivity `json:"review_activity"`
}

type LanguageCount struct {
	Language string `json:"language"`
	Count    int    `json:"count"`
}

func (r *documentRepo) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return count, nil
}

func (r *documentRepo) CountByLanguage(ctx context.Context) ([]LanguageCount, error) {
	query := `SELECT COALESCE(language, 'unknown'), COUNT(*) FROM documents GROUP BY language ORDER BY COUNT(*) DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("count by language: %w", err)
	}
	defer rows.Close()

	var counts []LanguageCount
	for rows.Next() {
		var lc LanguageCount
		if err := rows.Scan(&lc.Language, &lc.Count); err != nil {
			return nil, fmt.Errorf("scan language count: %w", err)
		}
		counts = append(counts, lc)
	}
	if counts == nil {
		counts = []LanguageCount{}
	}
	return counts, nil
}

func (r *chapterRepo) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_chapters`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count chapters: %w", err)
	}
	return count, nil
}

func (r *wordRepo) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM saved_words`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count words: %w", err)
	}
	return count, nil
}
