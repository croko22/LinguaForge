-- 004_srs.sql: Add word_reviews table for spaced repetition
CREATE TABLE IF NOT EXISTS word_reviews (
    id TEXT PRIMARY KEY,
    word_id TEXT NOT NULL UNIQUE REFERENCES saved_words(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'new',
    ease_factor REAL NOT NULL DEFAULT 2.5,
    interval_days INTEGER NOT NULL DEFAULT 0,
    repetitions INTEGER NOT NULL DEFAULT 0,
    lapses INTEGER NOT NULL DEFAULT 0,
    next_review TEXT NOT NULL DEFAULT (datetime('now')),
    last_reviewed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_word_reviews_next_review ON word_reviews(next_review);
CREATE INDEX IF NOT EXISTS idx_word_reviews_status ON word_reviews(status);
CREATE INDEX IF NOT EXISTS idx_word_reviews_word_id ON word_reviews(word_id);
