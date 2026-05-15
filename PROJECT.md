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
| Auth | Kinde (Google + GitHub OAuth) |
| Background jobs | goroutines (MVP) |

## Commands

Run all commands from the `backend/` directory.

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
go test ./backend/internal/... -run TestName

# Tidy dependencies
go mod tidy
```

## Backend Architecture

Clean architecture with strict layer separation. Dependencies flow inward only:

```
handler → usecase → repository
                 → domain
```

- `backend/internal/domain/` — entities and domain types, no external dependencies
- `backend/internal/usecase/` — business logic, depends only on domain and repository interfaces
- `backend/internal/repository/` — GORM implementations of repository interfaces defined in usecase
- `backend/internal/handler/` — Echo HTTP handlers, calls usecases only
- `backend/cmd/server/main.go` — entry point, wires everything together

## Key Domain Concepts

- **User** — authenticated user; holds `vision_enabled` flag to opt into AI organising. Defined in `backend/internal/domain/user.go`.
- **Image** — uploaded asset with metadata (path in R2, thumbnail path, folder, MIME type, Vision labels). Images are stored under `users/{kindeID}/images/` in the app's shared R2 bucket. `AILabels` stores the raw Vision API response and is persisted for future use.
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

## Frontend Architecture

The frontend uses **shadcn** components (`src/components/ui/`) as the only interface to UI primitives. shadcn in this project wraps **Base UI** (`@base-ui/react`), not Radix UI.

**Important:** Base UI and Radix UI share similar component names but have different prop APIs. Never copy props from Radix UI docs and apply them to shadcn components here — they will silently do nothing.

Key difference that has caused bugs:
- Radix UI `ContextMenuItem` uses `onSelect` to handle item clicks
- Base UI `ContextMenu.Item` uses `onClick`

Always use `onClick` on `ContextMenuItem`. When in doubt, check the wrapper in `src/components/ui/context-menu.tsx` to see which Base UI primitive it delegates to, then consult the Base UI docs for that primitive's props.

All direct `@base-ui/react` and `@radix-ui/*` imports must stay inside `src/components/ui/`. Application code imports only from `@/components/ui/`.

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP listen port |
| `R2_ACCOUNT_ID` | — | Cloudflare account ID; used to construct the R2 endpoint |
| `R2_ACCESS_KEY_ID` | — | R2 API token key ID |
| `R2_SECRET_ACCESS_KEY` | — | R2 API token secret; only shown once on creation |
| `R2_BUCKET_NAME` | — | R2 bucket name |
| `R2_PUBLIC_URL` | — | Public bucket URL (e.g. `https://pub-xxxx.r2.dev`); used for thumbnail CDN links — requires public access enabled on the bucket |
| `GOOGLE_VISION_API_KEY` | — | Google Vision API key for AI organising |
| `KINDE_ISSUER_URL` | — | Kinde domain (e.g. `https://yourapp.kinde.com`) |
| `KINDE_CLIENT_ID` | — | Kinde backend application client ID |
| `KINDE_CLIENT_SECRET` | — | Kinde backend application client secret |
| `KINDE_AUDIENCE` | — | API audience identifier registered in Kinde |
