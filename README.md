# LinguaForge

Language learning backend — ingest EPUB files, extract chapters, store in SQLite, serve via REST API.

Inspired by LingQ and Lute, but in Go. Faster, single binary, no Python overhead.

## Quick start

```bash
# Prerequisites: Go 1.22+
git clone https://github.com/croko22/LinguaForge.git
cd LinguaForge
go run ./cmd/api/
```

```bash
# Upload an EPUB
curl -X POST http://localhost:8080/api/documents \
  -F "file=@/path/to/book.epub"

# List documents
curl http://localhost:8080/api/documents

# Get chapters
curl http://localhost:8080/api/documents/<id>/chapters

# Get chapter content
curl http://localhost:8080/api/documents/<id>/chapters/0
```

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/documents` | Upload EPUB (multipart, field: `file`) |
| GET | `/api/documents` | List all documents |
| GET | `/api/documents/{id}` | Get document metadata |
| GET | `/api/documents/{id}/chapters` | List chapters (summary) |
| GET | `/api/documents/{id}/chapters/{index}` | Get chapter content |
| GET | `/health` | Health check |

## Configuration (env vars)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `./language-app.db` | SQLite database path |
| `UPLOAD_DIR` | `./uploads` | File storage directory |
| `MAX_UPLOAD_SIZE` | `52428800` (50MB) | Max file size in bytes |

## Project structure

```
cmd/api/main.go          # Entry point, wiring
internal/
  config/config.go       # Env-based config
  model/                 # Domain structs
  handler/               # HTTP handlers (chi router)
  service/               # Business logic
  parser/                # EPUB parser (zip + goquery)
  storage/               # File storage interface + local fs
  repository/            # SQLite CRUD
migrations/              # SQL schema
```

## Tech

- **Go** 1.22+ — single binary, no runtime deps
- **chi/v5** — lightweight HTTP router
- **SQLite** (modernc.org/sqlite) — pure Go, no CGO
- **EPUB parsing** — DIY with archive/zip + goquery (zero external EPUB libs)

## Testing

```bash
go test ./... -v
```

## Future

- PDF ingestion
- Spaced Repetition System (SRS)
- User auth & multi-user
- Reading UI
- Full-text search
