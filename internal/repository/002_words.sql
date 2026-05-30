CREATE TABLE IF NOT EXISTS saved_words (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL,
    word TEXT NOT NULL,
    translation TEXT NOT NULL DEFAULT '',
    source_lang TEXT NOT NULL DEFAULT 'en',
    target_lang TEXT NOT NULL DEFAULT 'es',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_saved_words_document ON saved_words(document_id);
CREATE INDEX IF NOT EXISTS idx_saved_words_created ON saved_words(created_at DESC);
