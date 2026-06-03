# Concurrent Book Processing

## Problem
`POST /api/documents` blocks HTTP request until EPUB/PDF is fully parsed (up to 30s timeout). Large books fail, UX is bad.

## Solution
In-memory worker pool. Upload returns immediately with `202 Accepted`. Worker processes in background.

## Architecture

### Upload flow (fast path)
1. Handler validates file type & size
2. Service creates DB record with `pending` status
3. Service stores file, enqueues job to worker pool
4. Returns `202` with doc ID immediately

### Processing flow (background)
1. Worker picks job from channel
2. Sets status to `processing`
3. Reads stored file, parses it
4. Writes chapters via `CreateBatch`
5. Sets status to `ready` (or `error` with message)

### Worker pool (`internal/worker/`)
- Fixed-size goroutine pool (configurable, default 2)
- Buffered channel for job queue (configurable, default 10)
- `Start()` launches workers, `Stop()` drains gracefully
- `Enqueue()` returns error when pool is stopped

### Service changes (`internal/service/`)
- `UploadDocument`: validate + create record + store + enqueue (returns doc with `pending`)
- `ProcessBook`: read stored file + parse + write chapters + set ready/error

### Handler changes (`internal/handler/`)
- `UploadDocument`: returns `202` + doc ID on success
- No other endpoint changes (list, get, chapters all work same)

## Testing
- Unit tests for worker pool (start/stop/enqueue/process)
- Unit tests for service (both paths)
- Handler test for 202 response
- Integration test: upload returns 202, poll until ready
