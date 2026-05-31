# PROJECT.md

This file provides guidance to agents working with code in this repository.
Refer to CONVENTIONS.md for coding conventions for backend, frontend, and extension.

## What is Bookleaf

Web-based image moodboarding app (inspired by Raindrop.io). A moodboard app with optional AI organising.

## Stack

Bookleaf has three layers: a Go + Echo backend, a React + Vite + 
TypeScript frontend, and a Chrome MV3 browser extension (Firefox-compatible). 
For exact dependencies, refer to backend/go.mod and frontend/package.json 
and extensions/package.json.

## Commands
Refer to Makefile for commands to run BE/FE/build extensions.
For local development/db migrations, refer to Makefile.local

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

## Frontend Architecture

The frontend uses shadcn components (`src/components/ui/`) as the only interface to UI primitives.

## Browser Extension

The extension lives under `/extensions`. It targets Chrome MV3 by default and is built with Firefox compatibility in mind.


## Environment Variables
Refer to .env.example for backend/frontend/extensions.