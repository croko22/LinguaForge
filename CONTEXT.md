# LinguaForge — Domain Glossary

## Frontend

- **Reader**: The main reading view where document text is displayed chapter by chapter. Desktop: text left, vocab panel right. Mobile: text full-width, vocab panel as slide-over overlay.
- **Reader Pagination**: Splits chapter content into pages that fill the viewport. Text container is `max-w-[68ch]` (~68 chars → ~11 words/line). Uses viewport height + lineHeight to estimate lines-per-page, then `lines * wordsPerLine = wordsPerPage`. Safety rails 30–1200. Recalculates on resize. Does NOT use Pretext or DOM measurement libs.
- **Reading Progress**: Persisted in DB (`reading_progress` table: document_id, chapter_index, percentage). localStorage used as instant cache for fast page restore. API: `PUT /api/documents/{id}/progress`, `GET /api/documents/{id}/progress`. Percentage = `(chapterIndex + 1) / totalChapters * 100`. Library cards show thin emerald progress bar.
- **Word Popover**: Inline tooltip that appears on clicking a word, showing translation, TTS button, and save action.
- **Word Panel**: Accumulates clicked words during a reading session. Desktop: right-side column. Mobile: slide-over overlay with backdrop.
- **Library**: Document list view — shows all uploaded documents with metadata.
- **Chapter Navigation**: Controls to move between chapters (prev/next buttons, chapter dropdown).

## Document

- **Document**: An uploaded file (EPUB) with metadata. Has a status lifecycle: pending → processing → ready (or error).
- **Chapter**: A logical section of a document. Has an index, title, and plain-text content.
- **Token**: A whitespace-delimited word in chapter content. Used for approximate word counting.

## Frontend stack

- **Framework**: React + Vite + Tailwind
- **Routing**: React Router DOM v7 — BrowserRouter
- **Word interaction**: Popover (inline) + Word Panel (side column)
- **Monorepo**: `frontend/` at repo root. Single `Makefile` for full-stack QA.
- **Dev**: `npm run dev` (Vite :5173) → proxies API to Go :8080
- **API client**: TanStack Query (React Query) + Vite proxy for dev CORS
- **State**: Server state via TanStack Query. UI state via React context or local state only.

## Status lifecycle

`pending → processing → ready` (or `error`)
