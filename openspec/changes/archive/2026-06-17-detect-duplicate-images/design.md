## Context

Bookleaf's upload flow is a three-step sequence driven by the frontend: `POST /images` (initiate) → presigned PUT to R2 → `POST /images/:id/complete` (commit). The backend never handles raw image bytes for new uploads — they go directly from the browser to R2.

Existing images already live in R2 with thumbnails but have no perceptual hash. New images can have their hash computed client-side during the upload flow (the browser already decodes the image to generate the thumbnail). Existing images require a server-side backfill that fetches thumbnails from R2.

## Goals / Non-Goals

**Goals:**
- Detect near-duplicate images (same content, different compression/resize/minor crop) at upload time
- Surface duplicates to the user via a warning toast (single upload) or inline badge (batch modal)
- Backfill pHash for all existing images via the periodic worker
- Store pHash in a format that enables efficient Hamming distance queries in Postgres

**Non-Goals:**
- Semantically similar image detection (ML territory)
- Blocking uploads on duplicate detection — upload always proceeds
- Cross-user duplicate detection
- Intra-batch duplicate detection (concurrent uploads make this unreliable; deferred)

## Decisions

### 1. FE computes pHash for new uploads; BE computes for backfill

The browser has the image in memory throughout the upload flow and already decodes it to generate a thumbnail. Computing pHash on the already-generated thumbnail blob costs one extra canvas read with no additional R2 round trip.

For existing images, the FE cannot retroactively compute hashes. The periodic worker fetches thumbnails from R2 and computes via `goimagehash` on the BE. This keeps `goimagehash` a BE-only dependency.

**Alternative considered**: BE computes hash for all images (including new uploads) by fetching from R2 post-commit. Rejected because it adds an extra R2 fetch per upload and requires async coordination to return duplicate info.

### 2. Storage: `bit(64)` column in Postgres

pHash produces a 64-bit value. `bit(64)` enables native Hamming distance via:

```sql
SELECT id, title, thumbnail_path
FROM images
WHERE user_id = $1
  AND id != $2
  AND phash IS NOT NULL
  AND bit_count(phash # $3::bit(64)) <= 10
```

`#` is bitwise XOR; `bit_count` counts set bits. This is the Hamming distance, fully evaluated in Postgres with no application-side loop.

The column is nullable — `NULL` means "not yet hashed" (existing images before backfill completes). Duplicate detection is silently skipped for un-hashed images.

**Alternative considered**: `bigint` — same 8 bytes, but XOR/popcount requires casting through `bit(64)` anyway, making queries more verbose. `bit(64)` is semantically cleaner.

**Alternative considered**: `text` (hex string) — no native distance function; would require fetching all rows and comparing in Go. Rejected.

In Go, `PHash` is `*string` with GORM tag `type:bit(64)`. The string is a 64-character binary literal (e.g. `"0110...1010"`). On insert, GORM passes it as a Postgres `bit(64)` literal.

### 3. Hamming distance threshold: 10

| Distance | Meaning |
|----------|---------|
| 0 | Identical content |
| 1–5 | Different JPEG quality, lossless re-save |
| 6–10 | Moderate resize, slight recompression |
| > 10 | Likely a different image |

Threshold of 10 catches the stated duplicate categories (different quality, resize, minor crop) without false positives on visually distinct images. Defined as a named constant in the usecase package.

### 4. FE pHash: inline DCT implementation on canvas pixels

The thumbnail generation already draws to a canvas via `createImageBitmap`. pHash computation adds:
1. Downsample to 32×32 greyscale on a canvas
2. 2D DCT over the pixel matrix
3. Take the top-left 8×8 of the DCT output (excluding DC component)
4. 64 output bits: each bit = (coefficient > mean)

This is ~50 lines of TypeScript with no new dependency. The result is a `bigint` formatted as a 64-character binary string for the API.

**Alternative considered**: `corona10/goimagehash`-compatible JS library — options are poorly maintained. Inline implementation is small and gives us direct control over the algorithm.

**Alternative considered**: Send raw image bytes to BE for hashing — requires BE to handle image bytes in the upload path, adding latency and complexity. Rejected.

### 5. Backfill: River periodic job, batch of 20 per tick, every 5 minutes

Follows the existing pattern for `CleanupStaleUploads` and `TrashPurge`. A new `BackfillPhashArgs{}` job is registered as a River periodic job running every 5 minutes.

Each tick:
1. `SELECT id, thumbnail_path FROM images WHERE phash IS NULL AND deleted_at IS NULL LIMIT 20`
2. For each: fetch thumbnail from R2 → compute pHash via `goimagehash.PerceptionHash` → `UPDATE images SET phash = ? WHERE id = ?`
3. Log count processed

At 20 images per tick, 12 ticks/hour: ~240 images/hour. For large accounts this is gradual, but duplicate detection degrades gracefully (silently skipped for un-hashed images). Once all images are hashed the job becomes a no-op.

**Alternative considered**: Enqueue one Asynq job per existing image at deployment — faster backfill but requires a seeding step and risks an R2 request spike. Rejected in favour of the self-healing sweep.

### 6. CompleteUpload response includes minimal duplicate info

```json
{
  "image_id": "...",
  "duplicates": [
    { "id": "...", "title": "...", "thumbnail_path": "..." }
  ]
}
```

Returns `id`, `title`, and `thumbnail_path` for each match. Title is used in the warning toast. Thumbnail path is included for future use (e.g. a side-by-side comparison modal). Empty array when no duplicates found.

## Risks / Trade-offs

**Concurrent batch uploads miss intra-batch duplicates** → Accepted for now. With `MAX_CONCURRENT = 3`, two duplicate files in the same batch may both complete before either is stored, so neither finds the other. Deferred; can be solved later by serialising uploads or doing a post-batch comparison pass.

**FE pHash computation adds latency to the upload flow** → The DCT over a 32×32 canvas is sub-millisecond in modern browsers. No user-facing impact expected. If profiling reveals otherwise, computation can be deferred to after the PUT but before `completeUpload`.

**Existing images backfill is slow for large accounts** → Gradual degradation is the trade-off. Duplicate detection for existing images improves over time. Batch size and interval can be tuned if needed without a code change (promote to config if necessary).

**pHash false positives near the threshold** → At threshold 10, visually distinct images with similar low-frequency content (e.g. two solid-colour images) could match. The warning is non-blocking (upload proceeds), so a false positive is a minor UX annoyance, not a data loss risk.

**Thumbnails missing for some images** → `ThumbnailPath` is nullable. The backfill sweep falls back to the full R2 path if thumbnail is nil. Computing pHash on the full image is slower but correct.

## Migration Plan

1. Deploy DB migration: adds `phash bit(64) NULL` to `images` — non-destructive, zero downtime.
2. Deploy app: new `BackfillPhashArgs` periodic job registers automatically; begins processing existing images in the background within 5 minutes.
3. New uploads immediately start receiving pHash detection on completion.
4. Backfill completes gradually; no manual intervention required.

**Rollback**: Remove the `phash` column via down migration. The FE `phash` field in the request body is ignored by the old handler — no FE rollback needed.
