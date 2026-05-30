package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func setupWordDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := RunMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestWordRepository_SaveAndList(t *testing.T) {
	db := setupWordDB(t)
	defer db.Close()

	repo := NewWordRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	word := &SavedWord{
		ID:          "w1",
		DocumentID:  "doc-1",
		Word:        "hello",
		Translation: "hola",
		SourceLang:  "en",
		TargetLang:  "es",
		CreatedAt:   now,
	}

	if err := repo.Save(ctx, word); err != nil {
		t.Fatalf("save: %v", err)
	}

	words, err := repo.ListByDocument(ctx, "doc-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Word != "hello" {
		t.Fatalf("expected 'hello', got '%s'", words[0].Word)
	}
	if words[0].Translation != "hola" {
		t.Fatalf("expected 'hola', got '%s'", words[0].Translation)
	}
}

func TestWordRepository_ListAll(t *testing.T) {
	db := setupWordDB(t)
	defer db.Close()

	repo := NewWordRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	repo.Save(ctx, &SavedWord{ID: "w1", DocumentID: "doc-1", Word: "hello", Translation: "hola", SourceLang: "en", TargetLang: "es", CreatedAt: now})
	repo.Save(ctx, &SavedWord{ID: "w2", DocumentID: "doc-1", Word: "world", Translation: "mundo", SourceLang: "en", TargetLang: "es", CreatedAt: now})

	words, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
}

func TestWordRepository_Delete(t *testing.T) {
	db := setupWordDB(t)
	defer db.Close()

	repo := NewWordRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	repo.Save(ctx, &SavedWord{ID: "w1", DocumentID: "doc-1", Word: "hello", Translation: "hola", SourceLang: "en", TargetLang: "es", CreatedAt: now})

	if err := repo.Delete(ctx, "w1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	words, _ := repo.ListAll(ctx)
	if len(words) != 0 {
		t.Fatalf("expected 0 words after delete, got %d", len(words))
	}
}
