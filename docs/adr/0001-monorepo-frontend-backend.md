# ADR 0001: Monorepo for frontend and backend

**Date**: 2026-05-25

## Status

Accepted

## Context

LinguaForge has a Go REST API backend and will have a React SPA frontend. We need to decide repository structure.

Options:
1. **Separate repos** (LinguaForge-API, LinguaForge-UI) — independent CI/CD, independent versioning
2. **Monorepo** (`frontend/` + `cmd/` + `internal/` under one roof) — single repo, shared CI

## Decision

Monorepo. Frontend lives at `frontend/` in the same repo as the Go backend.

## Rationale

- Single PR for full-stack changes (common in a 1-dev project)
- Single `make qa` runs both lint+test suites
- Shared tooling config at root (pre-commit, editorconfig)
- Easy to add e2e tests that cover both layers
- No cross-repo version synchronization overhead

## Consequences

- Root `Makefile` needs to delegate to `frontend/` targets
- CI must cache both Go modules and npm dependencies
- Frontend build artifacts should be gitignored
- If the team grows >3 devs, consider splitting
