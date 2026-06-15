# LinguaForge — AGENTS.md

## Entrypoint
- `cmd/api/main.go` — wires config → SQLite → repos → storage → parser → service → handlers → chi router
- Run: `go run ./cmd/api/`
- Config via env: `PORT`, `DB_PATH`, `UPLOAD_DIR`, `MAX_UPLOAD_SIZE`

## Architecture
- Deep modules: `DocumentIngester` (upload, process, delete), `DocumentReader` (cover, progress)
- Handler calls repos directly for trivial reads (list, get by ID, chapters, progress get)
- Shallow service pass-throughs were removed; handler ↔ repo for reads
- Parser is an interface (`internal/parser/parser.go`) — new formats implement `CanParse` + `Parse`
- Storage is an interface (`internal/storage/storage.go`) — currently `LocalFileStorage`, swappable for S3
- Repos are interfaces (`internal/repository/document.go`, `chapter.go`) — SQLite impl uses `modernc.org/sqlite` (pure Go, no CGO)

## QA (run before push)
```sh
make qa     # lint → sec → test (all)
make lint   # golangci-lint (12 linters, gocyclo≤15, gocognit≤20)
make sec    # gosec
make test   # go test -v ./...
```

pre-commit hooks: trailing-whitespace, eof-fixer, check-yaml, large-files (1MB), `go fmt`, `go vet`, `go test -short`

## Testing quirks
- `go test ./...` includes both unit (`internal/parser/parser_test.go`) and integration (`internal/integration_test.go`)
- Parser tests build EPUB in-memory via `archive/zip` — no fixture files on disk
- Integration test uses `httptest.Server` + temp SQLite + temp upload dir — no external services
- DB uses `modernc.org/sqlite` — pure Go, SQLite runs in-process, no libsqlite3 dependency

## Key interfaces
| Interface | Package | Methods |
|-----------|---------|---------|
| `Parser` | `parser` | `CanParse(filename) bool`, `Parse(ReaderAt, size)` |
| `FileStorage` | `storage` | `Store(ctx, filename, reader)`, `Get(ctx, path)`, `Delete(ctx, path)` |
| `DocumentRepository` | `repository` | `Create`, `GetByID`, `List`, `UpdateStatus`, `UpdateMetadata`, `Delete` |
| `ChapterRepository` | `repository` | `Create`, `CreateBatch`, `ListByDocumentID`, `GetByDocumentAndIndex`, `DeleteByDocumentID`, `CountByDocumentID` |
| `WordRepository` | `repository` | `Save`, `ListByDocument`, `ListAll`, `Delete`, `DeleteByDocumentID` |
| `ReviewRepository` | `repository` | `Create`, `GetByWordID`, `GetDueWords`, `UpdateReview`, `CountDue`, `DeleteByDocumentID` |
| `ReadingProgressRepository` | `repository` | `Upsert`, `GetByDocumentID`, `DeleteByDocumentID` |

## Epub parsing pipeline
`zip → container.xml → OPF → title/author/language → manifest → spine → NCX navPoints → chapter titles → read XHTML per spine item → goquery extract body text → collapse whitespace → count tokens`

Fallbacks: NCX → spine → single chapter.

## Entity status lifecycle
`pending → processing → ready` (or `error`)

## Module
`github.com/croko/language-app` — Go 1.26.3

## Dependencies
- `chi/v5` — HTTP router
- `modernc.org/sqlite` — pure-Go SQLite driver
- `PuerkitoBio/goquery` — HTML/XHTML parsing (epub chapter text extraction)
- `caarlos0/env/v11` — env var config
- `google/uuid` — ID generation
- `golang.org/x/net` — charset detection for non-UTF-8 XHTML

## Dev workflow
- New deep operation: handler method → service method → repo method(s)
- New trivial read: handler method → repo method directly
- New file format: implement `Parser` interface, add to `main.go` wiring
- New storage backend: implement `FileStorage` interface, swap in `main.go`
