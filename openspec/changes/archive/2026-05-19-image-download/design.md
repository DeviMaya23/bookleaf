## Context

Images are stored as private objects in Cloudflare R2. The existing `StorageService` interface exposes `GeneratePresignedGetURL` (used for thumbnail access) but does not support injecting response headers into the presigned URL. The image record already holds `r2_path` (the full-res object key) and `title` (usable as the download filename).

The download flow must be authenticated (user must own the image), return a usable URL quickly, and result in the browser saving the file rather than navigating to it.

## Goals / Non-Goals

**Goals:**
- Expose `GET /images/:id/download` returning a short-lived presigned URL the browser can use to download the full-res image
- Ensure the download triggers a file-save dialog (not inline display) via `Content-Disposition: attachment`
- Ownership check: only the image owner can request a download URL

**Non-Goals:**
- Streaming the image bytes through the API server (unnecessary — R2 presigned URLs handle this directly)
- Supporting batch or multi-image downloads
- Logging or tracking download events

## Decisions

### 1. Response shape: JSON with `download_url` (not a redirect)

Return `200 OK` with `{ "download_url": "..." }` rather than a `302` redirect.

**Why**: A redirect works in the browser but breaks non-browser clients and is inconsistent with the rest of the API's JSON conventions. The frontend can trigger the download via `window.open(url)` or an `<a download>` element.

**Alternative considered**: `302 Found` to the presigned URL — simpler but breaks API consistency and non-browser usage.

### 2. Extend `StorageService` with `GeneratePresignedDownloadURL`

Add `GeneratePresignedDownloadURL(ctx context.Context, key, filename string, ttl time.Duration) (string, error)` to the `StorageService` interface.

**Why**: R2 (S3-compatible) presigned GET URLs support `response-content-disposition` as a query parameter, which instructs the browser to treat the response as an attachment with a given filename. The existing `GeneratePresignedGetURL` does not accept this parameter. A dedicated method keeps the interface clean and avoids adding optional parameters to an existing method.

**Alternative considered**: Reuse `GeneratePresignedGetURL` and return the URL with no `Content-Disposition` — simpler but requires the frontend to handle download behaviour (e.g., `<a download>`), which can fail cross-origin or with some browsers.

### 3. TTL: 5 minutes

The presigned download URL SHALL expire after 5 minutes.

**Why**: Download URLs are consumed immediately after being requested. A short TTL reduces the window for URL leakage. 5 minutes gives enough headroom for slow connections without lingering indefinitely like the 24h thumbnail URLs.

### 4. Filename derived from image title

The `Content-Disposition` filename SHALL be derived from the image's `title` field with the correct extension appended based on `mime_type` (e.g., `title + ".jpg"`).

**Why**: `title` is the user-facing name of the image and maps cleanly to a meaningful download filename.

## Risks / Trade-offs

- [URL sharing] A user could share the download URL before it expires (5-min window). → Acceptable given the short TTL and that the full-res image is already their own asset.
- [StorageService interface change] Adding `GeneratePresignedDownloadURL` requires updating the mock in tests. → Low impact; follows the same pattern as existing methods.

## Migration Plan

No database migrations required. This change adds a new route and a new method on an existing interface only.

Deployment: standard — no feature flags, no rollout steps. Roll back by removing the route.
