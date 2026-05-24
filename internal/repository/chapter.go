package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/croko/language-app/internal/model"
)

// ChapterRepository defines persistence operations for chapters.
type ChapterRepository interface {
	Create(ctx context.Context, chapter *model.Chapter) error
	CreateBatch(ctx context.Context, chapters []*model.Chapter) error
	ListByDocumentID(ctx context.Context, documentID string) ([]*model.Chapter, error)
	GetByDocumentAndIndex(ctx context.Context, documentID string, index int) (*model.Chapter, error)
	DeleteByDocumentID(ctx context.Context, documentID string) error
	CountByDocumentID(ctx context.Context, documentID string) (int, error)
}

type chapterRepo struct {
	db *sql.DB
}

// NewChapterRepository creates a new ChapterRepository backed by SQLite.
func NewChapterRepository(db *sql.DB) ChapterRepository {
	return &chapterRepo{db: db}
}

func (r *chapterRepo) Create(ctx context.Context, chapter *model.Chapter) error {
	query := `
		INSERT INTO document_chapters (id, document_id, chapter_index, chapter_title, content, token_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		chapter.ID,
		chapter.DocumentID,
		chapter.ChapterIndex,
		chapter.ChapterTitle,
		chapter.Content,
		chapter.TokenCount,
		chapter.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create chapter: %w", err)
	}
	return nil
}

func (r *chapterRepo) CreateBatch(ctx context.Context, chapters []*model.Chapter) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op if already committed

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO document_chapters (id, document_id, chapter_index, chapter_title, content, token_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch insert: %w", err)
	}
	defer stmt.Close()

	for _, ch := range chapters {
		if _, err := stmt.ExecContext(ctx,
			ch.ID,
			ch.DocumentID,
			ch.ChapterIndex,
			ch.ChapterTitle,
			ch.Content,
			ch.TokenCount,
			ch.CreatedAt,
		); err != nil {
			return fmt.Errorf("insert chapter %d: %w", ch.ChapterIndex, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit batch insert: %w", err)
	}
	return nil
}

func (r *chapterRepo) ListByDocumentID(ctx context.Context, documentID string) ([]*model.Chapter, error) {
	query := `
		SELECT id, document_id, chapter_index, chapter_title, token_count, created_at
		FROM document_chapters
		WHERE document_id = ?
		ORDER BY chapter_index ASC
	`
	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []*model.Chapter
	for rows.Next() {
		ch := &model.Chapter{}
		var createdRaw any
		if err := rows.Scan(
			&ch.ID,
			&ch.DocumentID,
			&ch.ChapterIndex,
			&ch.ChapterTitle,
			&ch.TokenCount,
			&createdRaw,
		); err != nil {
			return nil, fmt.Errorf("list chapters scan: %w", err)
		}
		createdAt, err := parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		ch.CreatedAt = createdAt
		chapters = append(chapters, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list chapters rows: %w", err)
	}
	return chapters, nil
}

func (r *chapterRepo) GetByDocumentAndIndex(ctx context.Context, documentID string, index int) (*model.Chapter, error) {
	query := `
		SELECT id, document_id, chapter_index, chapter_title, content, token_count, created_at
		FROM document_chapters
		WHERE document_id = ? AND chapter_index = ?
	`
	ch := &model.Chapter{}
	var createdRaw any
	err := r.db.QueryRowContext(ctx, query, documentID, index).Scan(
		&ch.ID,
		&ch.DocumentID,
		&ch.ChapterIndex,
		&ch.ChapterTitle,
		&ch.Content,
		&ch.TokenCount,
		&createdRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chapter not found: %w", err)
		}
		return nil, fmt.Errorf("get chapter by document and index: %w", err)
	}
	createdAt, err := parseDBTime(createdRaw)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	ch.CreatedAt = createdAt
	return ch, nil
}

func (r *chapterRepo) DeleteByDocumentID(ctx context.Context, documentID string) error {
	query := `DELETE FROM document_chapters WHERE document_id = ?`
	result, err := r.db.ExecContext(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("delete chapters: %w", err)
	}
	_, _ = result.RowsAffected() // not needed, success is sufficient
	return nil
}

func (r *chapterRepo) CountByDocumentID(ctx context.Context, documentID string) (int, error) {
	query := `SELECT COUNT(*) FROM document_chapters WHERE document_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, documentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count chapters: %w", err)
	}
	return count, nil
}
