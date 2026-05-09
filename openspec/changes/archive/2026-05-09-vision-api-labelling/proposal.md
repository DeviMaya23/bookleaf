## Why

The `ai_labels` field on `Image` exists but is never populated. Integrating Google Vision API allows the app to auto-label uploaded images, enabling intelligent folder suggestions during the upload flow for users who opt into AI organising.

## What Changes

- Add a `VisionService` interface and Google Vision API implementation in a new `internal/vision/` package; send thumbnail bytes as base64 to the Vision API and return scored labels
- Add `GOOGLE_VISION_API_KEY` as an optional config env var under a new `Vision VisionConfig` sub-struct
- Restructure `CompleteUpload` in the image usecase: move R2 GetObject and thumbnail generation out of the background goroutine so they are synchronous; the goroutine now handles only the R2 PutObject and `UpdateThumbnailPath`
- After thumbnail generation, check `user.VisionEnabled`; if true, call Vision API with thumbnail bytes as base64, then persist the returned labels to `image.ai_labels`
- Change `CompleteUpload` handler response from `204 No Content` to `200 OK` with a body containing `image_id`, a single `folder_suggestion` object (the highest-scoring label), and an optional `warning string` field; `folder_suggestion` is `null` when Vision is not enabled or returns no labels
- If Vision API call fails or times out, the endpoint still returns `200 OK` with a `warning` field; failed Vision calls are not retried or stored for later

## Capabilities

### New Capabilities

- `vision-api-labelling`: Google Vision API client interface and implementation, label persistence on the `Image` record, and folder suggestion resolution — the highest-scoring label is matched case-insensitively against the user's existing folders; the suggestion includes `folder_id` (null if no match), `folder_name`, and `is_new` (true if no existing folder matched)

### Modified Capabilities

- `image-endpoints`: `POST /images/:id/complete` response changes from `204 No Content` to `200 OK` with `folder_suggestion` and optional `warning` fields
- `image-thumbnail`: goroutine for thumbnail handling is restructured — R2 GetObject and thumbnail generation are now synchronous steps in `CompleteUpload`; the goroutine handles only R2 PutObject and `UpdateThumbnailPath`
- `app-config`: new optional `GOOGLE_VISION_API_KEY` env var added under `Vision VisionConfig` sub-struct

## Impact

- `internal/vision/` — new package (interface + Google Vision HTTP client)
- `internal/config/config.go` — new `VisionConfig` sub-struct and `GOOGLE_VISION_API_KEY` optional env var
- `internal/usecase/image_usecase.go` — `CompleteUpload` and `uploadThumbnail` restructured; `VisionService` and `FolderRepository` dependencies added for label lookup and folder name matching
- `internal/handler/image.go` — `CompleteUpload` handler response type changed
- `internal/domain/image.go` — no change (AILabels field already exists)
- `internal/domain/user.go` — no change (VisionEnabled field already exists)
- New dependency: Google Cloud Vision API (HTTP/REST, no SDK required if using API key auth)
