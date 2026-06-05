package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/croko/language-app/internal/model"
	"github.com/google/uuid"
)

func setupReviewDB(t *testing.T) *sql.DB {
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

func seedWord(t *testing.T, db *sql.DB, id, documentID, word, translation string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO saved_words (id, document_id, word, translation, source_lang, target_lang, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`, id, documentID, word, translation, "en", "es")
	if err != nil {
		t.Fatalf("seed word: %v", err)
	}
}

func TestReviewRepository_CreateAndGetByWordID(t *testing.T) {
	db := setupReviewDB(t)
	defer db.Close()

	repo := NewReviewRepository(db)
	ctx := context.Background()

	// Seed a word first (FK constraint)
	seedWord(t, db, "word-1", "doc-1", "hello", "hola")

	now := time.Now().UTC().Truncate(time.Second)
	card := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       "word-1",
		Status:       model.ReviewStatusNew,
		EaseFactor:   2.5,
		IntervalDays: 0,
		Repetitions:  0,
		Lapses:       0,
		NextReview:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.Create(ctx, card); err != nil {
		t.Fatalf("create review card: %v", err)
	}

	got, err := repo.GetByWordID(ctx, "word-1")
	if err != nil {
		t.Fatalf("get by word id: %v", err)
	}
	if got == nil {
		t.Fatal("expected review card, got nil")
	}
	if got.WordID != "word-1" {
		t.Errorf("expected word_id=word-1, got %s", got.WordID)
	}
	if got.Status != model.ReviewStatusNew {
		t.Errorf("expected status=new, got %s", got.Status)
	}
	if got.EaseFactor != 2.5 {
		t.Errorf("expected ease_factor=2.5, got %f", got.EaseFactor)
	}
	if got.Repetitions != 0 {
		t.Errorf("expected repetitions=0, got %d", got.Repetitions)
	}
	if got.IntervalDays != 0 {
		t.Errorf("expected interval_days=0, got %d", got.IntervalDays)
	}
	if got.Lapses != 0 {
		t.Errorf("expected lapses=0, got %d", got.Lapses)
	}
}

func TestReviewRepository_GetDueWords(t *testing.T) {
	db := setupReviewDB(t)
	defer db.Close()

	repo := NewReviewRepository(db)
	ctx := context.Background()

	// Seed two words
	seedWord(t, db, "word-due", "doc-1", "hello", "hola")
	seedWord(t, db, "word-future", "doc-1", "world", "mundo")

	now := time.Now().UTC().Truncate(time.Second)

	// Card due now (next_review = now)
	dueCard := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       "word-due",
		Status:       model.ReviewStatusReview,
		EaseFactor:   2.5,
		IntervalDays: 1,
		Repetitions:  1,
		Lapses:       0,
		NextReview:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Create(ctx, dueCard); err != nil {
		t.Fatalf("create due card: %v", err)
	}

	// Card due in the future (next_review = now + 7 days)
	futureCard := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       "word-future",
		Status:       model.ReviewStatusReview,
		EaseFactor:   2.5,
		IntervalDays: 7,
		Repetitions:  2,
		Lapses:       0,
		NextReview:   now.AddDate(0, 0, 7),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Create(ctx, futureCard); err != nil {
		t.Fatalf("create future card: %v", err)
	}

	dueWords, err := repo.GetDueWords(ctx)
	if err != nil {
		t.Fatalf("get due words: %v", err)
	}

	if len(dueWords) != 1 {
		t.Fatalf("expected 1 due word, got %d", len(dueWords))
	}
	if dueWords[0].WordID != "word-due" {
		t.Errorf("expected word-due, got %s", dueWords[0].WordID)
	}
}

func TestReviewRepository_CountDue(t *testing.T) {
	db := setupReviewDB(t)
	defer db.Close()

	repo := NewReviewRepository(db)
	ctx := context.Background()

	// Seed words
	seedWord(t, db, "word-a", "doc-1", "hello", "hola")
	seedWord(t, db, "word-b", "doc-1", "world", "mundo")
	seedWord(t, db, "word-c", "doc-1", "foo", "bar")

	now := time.Now().UTC().Truncate(time.Second)

	// Two cards due now
	for i, wid := range []string{"word-a", "word-b"} {
		card := &model.ReviewCard{
			ID:           uuid.New().String(),
			WordID:       wid,
			Status:       model.ReviewStatusReview,
			EaseFactor:   2.5,
			IntervalDays: 1,
			Repetitions:  1,
			Lapses:       0,
			NextReview:   now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := repo.Create(ctx, card); err != nil {
			t.Fatalf("create card %d: %v", i, err)
		}
	}

	// One card in the future
	futureCard := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       "word-c",
		Status:       model.ReviewStatusReview,
		EaseFactor:   2.5,
		IntervalDays: 7,
		Repetitions:  2,
		Lapses:       0,
		NextReview:   now.AddDate(0, 0, 7),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Create(ctx, futureCard); err != nil {
		t.Fatalf("create future card: %v", err)
	}

	count, err := repo.CountDue(ctx)
	if err != nil {
		t.Fatalf("count due: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count=2, got %d", count)
	}
}

func TestReviewRepository_UpdateReview(t *testing.T) {
	db := setupReviewDB(t)
	defer db.Close()

	repo := NewReviewRepository(db)
	ctx := context.Background()

	seedWord(t, db, "word-upd", "doc-1", "hello", "hola")

	now := time.Now().UTC().Truncate(time.Second)

	card := &model.ReviewCard{
		ID:           uuid.New().String(),
		WordID:       "word-upd",
		Status:       model.ReviewStatusNew,
		EaseFactor:   2.5,
		IntervalDays: 0,
		Repetitions:  0,
		Lapses:       0,
		NextReview:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repo.Create(ctx, card); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Update the card after a review
	updatedCard := *card
	updatedCard.Status = model.ReviewStatusReview
	updatedCard.EaseFactor = 2.36
	updatedCard.IntervalDays = 1
	updatedCard.Repetitions = 1
	updatedCard.NextReview = now.AddDate(0, 0, 1)
	updatedCard.UpdatedAt = now.Add(time.Minute)
	now2 := now.Add(time.Minute)
	updatedCard.LastReviewedAt = &now2

	if err := repo.UpdateReview(ctx, &updatedCard); err != nil {
		t.Fatalf("update review: %v", err)
	}

	got, err := repo.GetByWordID(ctx, "word-upd")
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.Status != model.ReviewStatusReview {
		t.Errorf("expected status=review, got %s", got.Status)
	}
	if got.Repetitions != 1 {
		t.Errorf("expected repetitions=1, got %d", got.Repetitions)
	}
	if got.EaseFactor != 2.36 {
		t.Errorf("expected ease_factor=2.36, got %f", got.EaseFactor)
	}
	if got.IntervalDays != 1 {
		t.Errorf("expected interval_days=1, got %d", got.IntervalDays)
	}
	if got.LastReviewedAt == nil {
		t.Fatal("expected last_reviewed_at to be set")
	}
}

func TestReviewRepository_Create_DuplicateWordID(t *testing.T) {
	db := setupReviewDB(t)
	defer db.Close()

	repo := NewReviewRepository(db)
	ctx := context.Background()

	seedWord(t, db, "word-dup", "doc-1", "hello", "hola")

	now := time.Now().UTC().Truncate(time.Second)

	card := &model.ReviewCard{
		ID:         uuid.New().String(),
		WordID:     "word-dup",
		Status:     model.ReviewStatusNew,
		EaseFactor: 2.5,
		NextReview: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.Create(ctx, card); err != nil {
		t.Fatalf("first create: %v", err)
	}

	// Creating a second review card for the same word should fail (UNIQUE constraint on word_id)
	card2 := &model.ReviewCard{
		ID:         uuid.New().String(),
		WordID:     "word-dup",
		Status:     model.ReviewStatusNew,
		EaseFactor: 2.5,
		NextReview: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.Create(ctx, card2); err == nil {
		t.Error("expected error for duplicate word_id, got nil")
	}
}

func TestReviewRepository_GetByWordID_NotFound(t *testing.T) {
	db := setupReviewDB(t)
	defer db.Close()

	repo := NewReviewRepository(db)
	ctx := context.Background()

	got, err := repo.GetByWordID(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent word, got nil")
	}
	if got != nil {
		t.Fatalf("expected nil card, got %+v", got)
	}
}
