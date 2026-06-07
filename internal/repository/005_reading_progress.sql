CREATE TABLE IF NOT EXISTS reading_progress (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  chapter_index INTEGER NOT NULL DEFAULT 0,
  percentage REAL NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(document_id)
);

CREATE INDEX IF NOT EXISTS idx_reading_progress_document_id ON reading_progress(document_id);
