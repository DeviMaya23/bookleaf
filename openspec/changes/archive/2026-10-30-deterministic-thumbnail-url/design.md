# Design

## Context

See proposal.md — Why for motivation.

Presigned URL generation today uses the standard `s3.PresignClient`, which incorporates `time.Now()` into the HMAC signature. Four thumbnail call sites (across `imageUsecase`, `folderImagesUsecase`, `trashUsecase`, `shareUsecase`) each call `store.GeneratePresignedGetURL(ctx, path, 24h)` and get a distinct URL every time.

The r2.go from Booklet has already been copied into `backend/internal/storage/r2.go`. It adds:
- `deterministicSigner` — a `v4.HTTPPresigner` that substitutes a fixed signing time
- `GenerateDeterministicPresignedGetURL` — uses `deterministicSigner` with `Truncate(24h)` window start
- `now func() time.Time` clock field on `r2Storage`
- Updated `PutObject` with a `cacheControl string` parameter

The remaining work is purely interface plumbing and call-site wiring.

## Goals / Non-Goals

**Goals:**
- Thumbnail URLs are stable within a 24 h UTC window, enabling browser caching
- `StorageService` interface and all mocks compile cleanly
- `PutObject` interface matches the updated implementation signature

**Non-Goals:**
- Making full-resolution image URLs deterministic (they are low-frequency, per-view fetches)
- Setting `Cache-Control` on existing thumbnail objects in R2 (no backfill; only applies to new uploads)
- Any frontend or extension changes

## Decisions

### 1. New method on `StorageService` interface, not a wrapper

**Decision:** Add `GenerateDeterministicPresignedGetURL` to the `StorageService` interface.

**Why:** Usecases depend on the `StorageService` interface, not the concrete `r2Storage`. Bypassing the interface (e.g., type-asserting to `*r2Storage`) would break the clean-architecture dependency rule. Adding the method to the interface keeps the boundary clean and makes the behavior testable through the mock.

**Alternative considered:** A separate `DeterministicStorageService` interface with only the new method, composed into usecases that need it. Rejected: adds unnecessary indirection for a single method on an already-small interface.

### 2. No TTL parameter on `GenerateDeterministicPresignedGetURL`

**Decision:** The method takes only `ctx` and `key`; window and expiry constants live in `r2.go`.

**Why:** The window/expiry pairing is a coupled invariant (expiry must be ≥ 2× window). Exposing them as parameters invites callers to break the invariant. The method is intentionally opinionated.

### 3. `PutObject` cacheControl as empty-string opt-out

**Decision:** Callers that do not need `Cache-Control` pass `""`. No overloading or optional pattern.

**Why:** Go has no optional parameters. The empty-string convention is already used in the existing codebase (e.g., `ContentLength` zero-value opt-out in `GeneratePresignedPutURL`). Keeps the interface simple.

**What needs updating:** `image_storage.go` interface, `mockStorageService` in `image_usecase_test.go`. The embedding-based mocks (`stubFolderImagesStorage`, `spyShareStorage`, `stubStorageService`) embed the interface and satisfy it at compile time without changes.

### 4. `thumbnailURL()` helpers switch to the new method; `presignedGetTTL` stays for full-res

**Decision:** Each usecase's `thumbnailURL()` helper (and the `folder_images_usecase` inline call) switches to `GenerateDeterministicPresignedGetURL`. Full-resolution image URL calls keep `GeneratePresignedGetURL`.

**Why:** Thumbnails are the high-frequency case rendered in grids — repeated identical fetches per page load. Full-resolution URLs are fetched once per user-initiated view; the caching benefit is low and the per-view freshness semantics are simpler to reason about.

## Risks / Trade-offs

- **Window boundary re-fetch** — At the UTC day rollover, the URL changes. Browsers that cached the old URL will re-fetch once. This is expected and acceptable (one re-fetch per day per image).
- **48 h expiry window** — A URL generated just before midnight remains valid for the next 24 h. If a thumbnail is deleted or replaced during that window, the old URL is still reachable in R2 until its expiry. Acceptable: thumbnails are not replaced in place today.
- **Clock skew** — `now` defaults to `time.Now` (wall clock). Server clock drift could produce a different window start than the client's local time, but this has no correctness impact — all signing happens server-side.

## Migration Plan

- No database changes, no data migration.
- No API contract changes — `thumbnail_url` remains an opaque string in all JSON responses.
- Deploy is a standard rolling update. Old 24 h non-deterministic URLs issued before deploy expire naturally. After deploy, all new API responses carry the stable deterministic URLs.
- Rollback: revert the code changes. Clients fall back to non-deterministic URLs with the same 24 h TTL as before.
