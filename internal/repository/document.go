package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/croko/language-app/internal/model"
)

// DocumentRepository defines persistence operations for documents.
type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, id string) (*model.Document, error)
	List(ctx context.Context) ([]*model.DocumentSummary, error)
	UpdateStatus(ctx context.Context, id, status string, errMsg ...string) error
	UpdateMetadata(ctx context.Context, doc *model.Document) error
}

type documentRepo struct {
	db *sql.DB
}

// NewDocumentRepository creates a new DocumentRepository backed by SQLite.
func NewDocumentRepository(db *sql.DB) DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) Create(ctx context.Context, doc *model.Document) error {
	query := `
		INSERT INTO documents (id, title, filename, file_type, file_size, storage_path, status, error_message, language, chapter_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		doc.ID,
		doc.Title,
		doc.Filename,
		doc.FileType,
		doc.FileSize,
		doc.StoragePath,
		doc.Status,
		nullIfEmpty(doc.ErrorMessage),
		doc.Language,
		doc.ChapterCount,
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

func (r *documentRepo) GetByID(ctx context.Context, id string) (*model.Document, error) {
	query := `
		SELECT id, title, filename, file_type, file_size, storage_path, status,
		       COALESCE(error_message, ''), COALESCE(language, ''),
		       chapter_count, created_at, updated_at
		FROM documents
		WHERE id = ?
	`
	doc := &model.Document{}
	var createdRaw any
	var updatedRaw any
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID,
		&doc.Title,
		&doc.Filename,
		&doc.FileType,
		&doc.FileSize,
		&doc.StoragePath,
		&doc.Status,
		&doc.ErrorMessage,
		&doc.Language,
		&doc.ChapterCount,
		&createdRaw,
		&updatedRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %w", err)
		}
		return nil, fmt.Errorf("get document by id: %w", err)
	}

	createdAt, err := parseDBTime(createdRaw)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := parseDBTime(updatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	doc.CreatedAt = createdAt
	doc.UpdatedAt = updatedAt
	return doc, nil
}

func (r *documentRepo) List(ctx context.Context) ([]*model.DocumentSummary, error) {
	query := `
		SELECT id, title, file_type, file_size, status,
		       COALESCE(language, ''), chapter_count, created_at
		FROM documents
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var summaries []*model.DocumentSummary
	for rows.Next() {
		s := &model.DocumentSummary{}
		var createdRaw any
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.FileType,
			&s.FileSize,
			&s.Status,
			&s.Language,
			&s.ChapterCount,
			&createdRaw,
		); err != nil {
			return nil, fmt.Errorf("list documents scan: %w", err)
		}
		createdAt, err := parseDBTime(createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		s.CreatedAt = createdAt
		summaries = append(summaries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list documents rows: %w", err)
	}
	return summaries, nil
}

func (r *documentRepo) UpdateStatus(ctx context.Context, id, status string, errMsg ...string) error {
	query := `UPDATE documents SET status = ?, updated_at = datetime('now')`
	args := []any{status}

	if len(errMsg) > 0 {
		query += ", error_message = ?"
		args = append(args, errMsg[0])
	}

	query += " WHERE id = ?"
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update document status rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *documentRepo) UpdateMetadata(ctx context.Context, doc *model.Document) error {
	query := `
		UPDATE documents
		SET title = ?, language = ?, chapter_count = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		doc.Title,
		doc.Language,
		doc.ChapterCount,
		doc.UpdatedAt,
		doc.ID,
	)
	if err != nil {
		return fmt.Errorf("update document metadata: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update document metadata rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("document not found: %w", sql.ErrNoRows)
	}
	return nil
}

// nullIfEmpty returns nil for empty strings so they're stored as SQL NULL.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
