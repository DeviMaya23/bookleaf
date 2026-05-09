## Context

`CompleteUpload` (`POST /images/:id/complete`) currently fires a background goroutine that fetches the original image from R2, generates a thumbnail, uploads it back to R2, and updates `thumbnail_path`. The response is `204 No Content`.

The `Image` domain already has `ai_labels jsonb` and `User` already has `vision_enabled bool`, both introduced in earlier migrations. Neither field is populated today.

Google Vision API will receive the thumbnail as a base64-encoded JPEG and return scored labels (e.g. `"Nature": 0.98`). The top label drives a single folder suggestion — matched case-insensitively against the user's existing folders.

## Goals / Non-Goals

**Goals:**
- Populate `ai_labels` on the image record for users with `vision_enabled = true`
- Return a single `folder_suggestion` object from `CompleteUpload` (highest-scoring label, with folder match resolution)
- Keep the endpoint non-blocking on Vision failures — always return `200 OK`, degrade gracefully with a `warning` field
- Add `GOOGLE_VISION_API_KEY` as an optional env var without breaking existing deploys

**Non-Goals:**
- Retrying or queuing failed Vision API calls
- Surfacing `ai_labels` via any other endpoint in this change
- Creating the suggested folder automatically
- Supporting Vision for users with `vision_enabled = false`

## Decisions

### 1. Move thumbnail generation out of the goroutine; goroutine handles only R2 upload

**Decision:** `GetObject` and `ThumbnailService.Generate` are called synchronously in `CompleteUpload` before the goroutine is fired. The goroutine receives the pre-generated thumbnail bytes and only calls `PutObject` + `UpdateThumbnailPath`.

**Rationale:** Vision API requires the thumbnail bytes. Since the goroutine is fire-and-forget with no return channel, making the thumbnail available to a synchronous Vision call requires it to be generated before the goroutine starts. The thumbnail key is deterministic (`users/{userID}/thumbnails/{imageID}.jpg`), so there is no correctness issue generating it before confirming the R2 upload.

**Alternative considered:** Keep everything in the goroutine and use a channel to pass results back to the handler. Rejected — it couples the async goroutine to the request lifecycle and complicates error handling.

**Trade-off:** For all users (not just `vision_enabled` ones), GetObject and thumbnail generation now block the `CompleteUpload` response. This adds latency to the endpoint. However, thumbnail generation is fast (in-process resize), and the R2 GetObject call was already happening — just moved earlier.

---

### 2. New `internal/vision/` package with interface + HTTP client (no SDK)

**Decision:** Define a `VisionService` interface in `internal/vision/` and implement it as a thin HTTP client calling the Google Vision REST API with an API key. No Google Cloud SDK.

```go
type VisionService interface {
    AnnotateImage(ctx context.Context, imageBytes []byte) ([]Label, error)
}

type Label struct {
    Description string
    Score       float32
}
```

**Rationale:** The Google Cloud Vision Go SDK pulls in a large dependency tree (gRPC, protobuf). Using the REST API with an API key is simpler, has no additional auth setup, and keeps the binary lean. The interface keeps the usecase testable without hitting the real API.

---

### 3. Folder name matching via a new `FindByName` method on `FolderRepository`

**Decision:** Add `FindByName(ctx context.Context, userID, name string) (*domain.Folder, error)` to `FolderRepository`. The SQL implementation uses `ILIKE` for a case-insensitive match. Returns `nil, nil` if no folder is found (not `ErrRecordNotFound`).

**Rationale:** The alternative — fetching all folders with `List` and filtering in Go — works at small scale but pushes a data-access concern into application logic. A targeted SQL query is cleaner and does not load unnecessary folder rows.

---

### 4. `CompleteUpload` usecase returns a result struct

**Decision:** Change the `ImageUsecase.CompleteUpload` signature from returning `error` to `(*CompleteUploadResult, error)`.

```go
type FolderSuggestion struct {
    FolderID   *uuid.UUID
    FolderName string
    IsNew      bool
}

type CompleteUploadResult struct {
    ImageID          uuid.UUID
    FolderSuggestion *FolderSuggestion // nil if vision not enabled or no labels returned
    Warning          string            // non-empty if Vision API failed
}
```

**Rationale:** The handler needs structured data to build the response body. Passing it through a result struct keeps the usecase boundary clean and avoids side-channel state.

---

### 5. Vision API call is gated and fails open

**Decision:** If `GOOGLE_VISION_API_KEY` is unset or empty, `VisionService` is not initialised and `CompleteUpload` skips Vision entirely (no error, no warning). If the key is set but the API call fails or times out, `CompleteUpload` logs the error, sets `Warning` on the result, and continues. A 5-second timeout is applied to the Vision API context.

**Rationale:** Vision is an enhancement, not a core upload flow requirement. Failing open ensures existing behaviour is preserved for all deploys without the env var set.

---

### 6. `CompleteUpload` delegates to two private helpers

**Decision:** The synchronous work inside `CompleteUpload` is split into two private methods on `imageUsecase`:

- `prepareThumbnail(ctx, image) ([]byte, error)` — fetches the original from R2 and generates thumbnail bytes
- `runVisionFlow(ctx, imageID, userID, thumbnailBytes) (suggestion *FolderSuggestion, warning string)` — checks `VisionEnabled`, calls Vision API, persists labels, resolves folder suggestion; returns no error (always fails open)

**Rationale:** Keeps `CompleteUpload` readable as an orchestration method. `runVisionFlow` in particular is isolated so that a v2 iteration (e.g. swapping Vision provider, adding async queuing, or changing the suggestion logic) has a clear, contained seam to cut against without touching thumbnail or upload logic.

---

### 7. `GOOGLE_VISION_API_KEY` is optional in config

**Decision:** Add a `VisionConfig` sub-struct to `Config` with a single `APIKey string` field, loaded via `envWithDefault("GOOGLE_VISION_API_KEY", "")`. Not required — `config.Load()` does not return an error if it is absent.

**Rationale:** The key may not be present in all environments (e.g. local dev, staging without Vision enabled). Making it required would break existing deploys.

## Risks / Trade-offs

- **Increased `CompleteUpload` latency for all users** — thumbnail GetObject + Generate now block the response. Mitigation: thumbnail generation is a fast in-process resize; the R2 GetObject was already happening in the goroutine anyway. Monitor p95 latency on this endpoint post-deploy.

- **Vision API adds a synchronous network call** — a slow or flaky Vision API response delays the endpoint for `vision_enabled` users. Mitigation: 5-second context timeout; Vision failure returns 200 with `warning` field.

- **Thumbnail bytes held in memory until goroutine starts** — the thumbnail `io.Reader` must be buffered into a `[]byte` so it can be passed to both Vision and the R2 upload goroutine. For a 600×600 JPEG this is typically < 100 KB. Acceptable.

- **`FolderRepository` interface grows** — adding `FindByName` changes the interface, requiring all mocks and test doubles to be updated. Low impact given the small number of callers.

## Migration Plan

1. Deploy with `GOOGLE_VISION_API_KEY` unset — all users see existing behaviour, `vision_enabled` is ignored, `folder_suggestion` is `null`
2. Set `GOOGLE_VISION_API_KEY` in the target environment
3. Enable `vision_enabled = true` for specific users via DB to validate end-to-end
4. No schema migrations required (both `ai_labels` and `vision_enabled` columns already exist)

**Rollback:** Unset `GOOGLE_VISION_API_KEY` — Vision is skipped, endpoint returns `null` for `folder_suggestion`. The `204 → 200` response change is the only breaking change for the FE; coordinate deploy timing.

## Open Questions

- Should `ai_labels` store all returned Vision labels (full list) or only the top label? Current assumption: store all labels returned by the API, use only the top one for `folder_suggestion`.
- Is there a Vision API quota / billing concern that needs a rate-limit guard before enabling broadly?
