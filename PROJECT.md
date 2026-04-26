# PROJECT.md

This file provides guidance to agents working with code in this repository.

## What is Bookleaf

Web-based image moodboarding app (inspired by Raindrop.io). Key differentiator: **BYOS** — users bring their own Cloudflare R2 bucket for storage, and optionally **BYOV** — their own Google Vision API key for AI auto-categorisation.

MVP scope: user registration, connect R2 bucket, upload images + thumbnail generation, browse gallery, manual folder management, optional Google Vision auto-categorisation.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go + Echo, clean architecture |
| Frontend | React + Vite + TypeScript + Tailwind + shadcn |
| Database | PostgreSQL + GORM |
| Storage | Cloudflare R2 (S3-compatible, user-supplied credentials) |
| Auth | Clerk (Google + GitHub OAuth) |
| Background jobs | goroutines (MVP) |

## Commands

```bash
# Run the server (defaults to port 8080)
go run ./cmd/server

# Run with custom port
PORT=3000 go run ./cmd/server

# Build
go build ./cmd/server

# Run tests
go test ./...

# Run a single test
go test ./internal/... -run TestName

# Tidy dependencies
go mod tidy
```

## Backend Architecture

Clean architecture with strict layer separation. Dependencies flow inward only:

```
handler → usecase → repository
                 → domain
```

- `internal/domain/` — entities and domain types, no external dependencies
- `internal/usecase/` — business logic, depends only on domain and repository interfaces
- `internal/repository/` — GORM implementations of repository interfaces defined in usecase
- `internal/handler/` — Echo HTTP handlers, calls usecases only
- `cmd/server/main.go` — entry point, wires everything together

## Key Domain Concepts

- **BucketConfig** — per-user Cloudflare R2 credentials (access key, secret, bucket name, endpoint). First-class domain concept; every storage operation is scoped to the authenticated user's own bucket.
- **Image** — uploaded asset with metadata (path in R2, thumbnail path, folder, MIME type, Vision labels if BYOV enabled). Metadata is stored in PostgreSQL.
- **Folder** — user-managed grouping of images, manual hierarchy.

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP listen port |
