# PROJECT.md

This file provides guidance to agents working with code in this repository.

## What is Bookleaf

Web-based image moodboarding app (inspired by Raindrop.io). A moodboard app with optional AI organising.

MVP scope: user registration, upload images + thumbnail generation, browse gallery, manual folder management, optional AI-assisted folder suggestions on upload.

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go + Echo, clean architecture |
| Frontend | React + Vite + TypeScript + Tailwind + shadcn |
| Database | PostgreSQL + GORM |
| Storage | Cloudflare R2 (S3-compatible, app-provided credentials) |
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

- **User** — authenticated user; holds `vision_enabled` flag to opt into AI organising.
- **Image** — uploaded asset with metadata (path in R2, thumbnail path, folder, MIME type, Vision labels). Images are stored under `users/{clerkID}/images/` in the app's shared R2 bucket. `AILabels` stores the raw Vision API response and is persisted for future use.
- **Folder** — user-managed grouping of images, manual hierarchy.

## AI Organising (folder suggestion)

When `vision_enabled` is true, the following happens synchronously on upload:

1. Call Google Vision API → returns labels with confidence scores.
2. Store labels as `AILabels` on the Image record.
3. Match labels against the user's existing folder names (case-insensitive).
   - If a match is found → suggest that folder.
   - If no match → suggest the highest-scoring label as a new folder name.
4. Suggestion is shown to the user in the upload UI (ephemeral, not persisted).
5. If the user accepts:
   - Existing folder → set `Image.FolderID`.
   - New folder name → create the folder, then set `Image.FolderID`.
6. If the user ignores → image remains unorganised.

If the user has no folders yet, step 3 always falls through to suggesting a new folder from the top label.

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP listen port |
| `R2_ACCESS_KEY_ID` | — | Cloudflare R2 access key |
| `R2_SECRET_ACCESS_KEY` | — | Cloudflare R2 secret key |
| `R2_BUCKET_NAME` | — | R2 bucket name |
| `R2_ENDPOINT_URL` | — | R2 endpoint (e.g. `https://<account>.r2.cloudflarestorage.com`) |
| `GOOGLE_VISION_API_KEY` | — | Google Vision API key for AI organising |
