-- Database schema v1: EPUB ingestion system
-- Applied via internal/repository/migrate.go

-- documents table: tracks uploaded EPUB files and their processing state
CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,  -- UUID
    title TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL DEFAULT 'epub',  -- 'epub' for now
    file_size INTEGER NOT NULL DEFAULT 0,
    storage_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',  -- 'pending', 'processing', 'ready', 'error'
    error_message TEXT,
    language TEXT DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at DESC);

-- document_chapters table: extracted chapters/segments per document
CREATE TABLE IF NOT EXISTS document_chapters (
    id TEXT PRIMARY KEY,  -- UUID
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chapter_index INTEGER NOT NULL,
    chapter_title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    token_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(document_id, chapter_index)
);

CREATE INDEX IF NOT EXISTS idx_chapters_document_id ON document_chapters(document_id);
CREATE INDEX IF NOT EXISTS idx_chapters_document_index ON document_chapters(document_id, chapter_index);
