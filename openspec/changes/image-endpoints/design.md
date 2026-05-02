## Context

The `Image` domain struct and `000003_create_images` migration already exist. Images support soft delete via `gorm.DeletedAt`. The project follows a handler → usecase → repository layering pattern (established by user and folder modules). R2 (Cloudflare object storage) is the designated store for full-resolution images and thumbnails; its S3-compatible API means the AWS SDK v2 can be used as the client.

Thumbnail generation has not been implemented yet. Google Vision AI label generation is explicitly out of scope for this change.

Note: The `Image` GORM struct tag for the folder FK still reads `OnDelete:SET NULL` — migration 000005 already changed the DB constraint to `ON DELETE RESTRICT`. The struct tag should be corrected to `OnDelete:RESTRICT` as part of this change to keep them in sync.

## Goals / Non-Goals

**Goals:**
- Two-step upload flow: backend generates presigned PUT URL, client uploads directly to R2, client notifies backend on completion
- Gallery list endpoint filterable by folder, returning thumbnail CDN URLs (no presigned URLs)
- Detail endpoint returning fresh 24-hour presigned GET URL for full-resolution image (never stored)
- Soft delete, trash list, and restore endpoints
- R2 storage service with presigned URL generation and CDN URL construction
- Server-side thumbnail generation (max 600×600) via goroutine after upload completion

**Non-Goals:**
- Google Vision AI label generation
- Client-side upload progress tracking
- Pagination on list/trash endpoints (flat list, same as folder module)
- Image deduplication or validation of R2 upload success

## Decisions

### D1: Two-step upload — presigned PUT URL, not server-side proxy

**Decision**: `POST /images` creates the DB record and returns a presigned PUT URL. The client uploads the binary directly to R2. `POST /images/:id/complete` signals the backend to start thumbnail generation.

**Why over server-side proxy**: Proxying image bytes through the backend wastes server memory and bandwidth for potentially large files. Presigned PUT is the standard pattern for direct-to-object-storage uploads and scales without backend changes.

**Tradeoff**: The backend trusts the client's `/complete` signal — it does not verify the object actually exists in R2 before generating a thumbnail. A goroutine that fetches a non-existent object will error silently. This is acceptable at MVP scale; a retry mechanism can be added later.

### D2: R2 object key scheme

**Decision**:
- Full images: `users/{kindeID}/images/{imageID}.{ext}`
- Thumbnails: `users/{kindeID}/thumbnails/{imageID}.jpg`

Extension is derived from the `mime_type` field at record creation time (`image/jpeg` → `.jpg`, `image/png` → `.png`, `image/webp` → `.webp`). Thumbnails are always JPEG regardless of original format.

**Why**: Namespacing by user ID keeps the bucket organised and makes user data deletion straightforward in future. Separate `images/` and `thumbnails/` prefixes allow independent CDN cache rules.

### D3: StorageService as an injectable interface

**Decision**: Define a `StorageService` interface in `internal/storage/` with methods `GeneratePresignedPutURL`, `GeneratePresignedGetURL`, `GetObject`, `PutObject`. The R2 implementation lives in the same package. The usecase receives `StorageService` as a constructor argument.

**Why**: Keeps the usecase testable without a real R2 bucket. The interface boundary also makes swapping storage providers straightforward.

### D4: Thumbnail generation in a goroutine, errors logged not surfaced

**Decision**: After `POST /images/:id/complete` returns `204`, the handler fires `go generateAndStoreThumbnail(...)`. Errors inside the goroutine are logged but not surfaced to the client.

**Why**: Thumbnail generation involves two R2 round-trips (GET original + PUT thumbnail) and image processing. Blocking the HTTP response on this would add significant latency. The thumbnail is a derived asset — the image is still accessible (via presigned GET) even if thumbnail generation fails. A failed thumbnail means `thumbnail_path` stays `nil`; clients should handle a null thumbnail URL gracefully.

### D5: Presigned GET URL generated fresh on every GET /images/:id — never stored

**Decision**: `GET /images/:id` calls `StorageService.GeneratePresignedGetURL` on each request with a 24-hour TTL. The URL is returned in the response body but never persisted to the DB.

**Why**: Storing presigned URLs creates stale-URL problems as they expire. On-demand generation is cheap (no network call, just a signature computation) and always returns a fresh URL. The 24-hour TTL is long enough for a typical browsing session.

### D6: Soft delete via GORM DeletedAt — no new migration needed

**Decision**: `DELETE /images/:id` relies on GORM's built-in soft delete (sets `deleted_at`). `GET /images/trash` uses `db.Unscoped().Where("deleted_at IS NOT NULL")`. The `images` table already has the `deleted_at` column and index from migration 000003.

**Why**: The domain struct already has `DeletedAt gorm.DeletedAt` — this is zero-cost to implement at the repository layer.

### D7: R2 config added to existing Config struct

**Decision**: Add `R2Config` struct to `internal/config/config.go` and load five new required env vars: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL`.

**Why over a separate config file**: Consistent with the existing single-file config pattern. All env var loading and validation is in one place.

## Risks / Trade-offs

- **Orphaned R2 objects**: If the DB record is created but the client never uploads (or calls `/complete`), the R2 key will be allocated in the presigned URL but no object will be stored. → Mitigation: acceptable at MVP; a cleanup job can be added later.
- **Silent thumbnail failures**: If thumbnail generation fails inside the goroutine, `thumbnail_path` stays nil and the client receives a null CDN URL. → Mitigation: log errors; clients must handle null thumbnail URLs gracefully.
- **No upload verification**: The backend does not verify the R2 object exists before marking the upload complete. → Mitigation: acceptable at MVP; could add a `HeadObject` check in the goroutine before thumbnail generation.
- **Goroutine leak on server shutdown**: In-flight goroutines generating thumbnails will be abandoned on graceful shutdown. → Mitigation: acceptable at MVP; use a `sync.WaitGroup` or context propagation if this becomes an issue.

## Migration Plan

1. Add `R2Config` to `internal/config/config.go` and update config tests
2. Correct `Image` GORM struct folder FK tag from `OnDelete:SET NULL` → `OnDelete:RESTRICT`
3. Add `internal/storage/r2.go` and `internal/thumbnail/thumbnail.go`
4. Add repository, usecase, handler
5. Wire into `main.go` and register routes
6. Set R2 env vars in `.env.local`
