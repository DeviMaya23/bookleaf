# Proposal

## Why

Every request for a thumbnail currently generates a freshly signed presigned URL, because the AWS v4 signing process uses the current timestamp. This means the `<img src>` URL changes on every API call, preventing the browser from caching thumbnail images and causing redundant R2 fetches on every page load.

## What Changes

- Add `GenerateDeterministicPresignedGetURL(ctx, key string) (string, error)` to the `StorageService` interface and its `r2Storage` implementation. This method signs with the start of the current UTC day (truncated to a 24 h window) rather than the current time, producing a stable URL for the full window. URLs expire after 48 h (2× the window), so a URL minted at the end of one window is still valid for the next.
- Add a `deterministicSigner` type that implements `v4.HTTPPresigner`, substituting a fixed signing time into the standard AWS v4 signer.
- Add a `now func() time.Time` clock field to `r2Storage` (set to `time.Now` in production) to make the window calculation testable.
- Update `PutObject` to accept an optional `cacheControl string` parameter. When non-empty, the `Cache-Control` header is set on the R2 object at upload time; R2 will echo it back in presigned GET responses, complementing the deterministic URL for browser caching.
- Switch all four thumbnail URL call sites (in `image_usecase`, `folder_images_usecase`, `trash_usecase`, `share_usecase`) from `GeneratePresignedGetURL` to `GenerateDeterministicPresignedGetURL`.
- Update the `StorageService` interface declaration, the mock in `image_usecase_test.go`, and any other mock that implements the interface.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `r2-storage`: Add `GenerateDeterministicPresignedGetURL` to the interface and implementation; update `PutObject` signature to include `cacheControl string`.
- `image-thumbnail`: Add a requirement that thumbnail GET URLs are produced by `GenerateDeterministicPresignedGetURL` so they are stable within a 24 h window and browser-cacheable.

## Impact

- **Backend** — `backend/internal/storage/r2.go`: new method, new signer type, `now` clock field, `PutObject` signature change.
- **Backend** — `backend/internal/usecase/image_storage.go`: interface updated (new method, `PutObject` sig).
- **Backend** — `backend/internal/usecase/image_usecase.go`, `folder_images_usecase.go`, `trash_usecase.go`, `share_usecase.go`: thumbnail URL call sites switched.
- **Backend** — `backend/internal/usecase/image_usecase_test.go`: mock updated for `PutObject` and new method stub added.
- No API contract changes visible to the frontend or extension (thumbnail URLs remain opaque strings in JSON responses).
- No database changes.
