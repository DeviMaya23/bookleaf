## Context

Thumbnail URLs are currently built via `CDNUrl` in the handler layer — a direct public bucket URL. The main image URL in `GetImage` already uses presigned GET URLs via the usecase, so the pattern is established. This change brings thumbnails in line with that pattern and removes the public bucket dependency entirely.

`PresignGetObject` from the AWS SDK v2 is a local operation — it computes an HMAC signature from the credentials without making a network call. Generating presigned URLs for a full image list page is purely CPU work and has no latency cost.

## Goals / Non-Goals

**Goals:**
- All object reads from R2 go through presigned GET URLs — no path remains that requires the bucket to be public
- Thumbnail URL generation belongs to the usecase layer — the handler has no storage knowledge for reads
- `StorageService` interface shrinks: `CDNUrl` is removed
- `R2_PUBLIC_URL` env var is removed; `ImageHandler` no longer holds a `store` reference

**Non-Goals:**
- Changing the presigned URL TTL (stays at 24h, consistent with `image_url`)
- Frontend changes (transparent — `thumbnail_url` field type and nullability are unchanged)
- Presigned URL caching (local signing is fast enough; caching adds complexity without meaningful gain)

## Decisions

**1. `thumbnailURL` is a private helper on `imageUsecase`, not on `StorageService`**

The nil-check, presign call, and error-swallow logic is identical across all 5 usecase methods. Centralising it avoids repetition and keeps the policy (nil on failure) in one place.
- Alternative: inline per method — repetitive, policy can drift.
- Alternative: add to `StorageService` — wrong layer; the interface should not know about degradation policy.

**2. Presign failure returns nil thumbnail, not an error**

Thumbnail presence is cosmetic — the image record and full image URL are unaffected. Failing an entire list request because a thumbnail URL could not be signed would be disproportionate.
- Alternative: propagate error — breaks the list for a non-critical field.

**3. Introduce `ImageItem{Image, ThumbnailURL}` as a shared usecase return type**

`ListImages`, `ListTrashed`, `Restore`, and `UpdateImage` all return image data that includes a thumbnail URL. A single shared type avoids four separate wrapper structs.
- Alternative: add `ThumbnailURL` field directly to `domain.Image` — domain structs should not carry computed/external URLs.
- Alternative: return `ThumbnailURL` as a separate parallel slice — error-prone to keep in sync.

**4. `toImageResponse` becomes a package-level function, not a method on `ImageHandler`**

After this change it has no storage dependency — it is a pure field mapping. A method receiver that is never used signals a wrong ownership. The handler no longer needs a `store` field at all for read paths.
- Alternative: keep as method — misleading; implies handler still has storage responsibility.

**5. `R2_PUBLIC_URL` becomes a hard error if still present — no, just removed**

The env var is simply dropped from `loadFromEnv`. If an operator still has it set, it is silently ignored (standard Go `os.Getenv` behaviour). No migration ceremony needed.

## Risks / Trade-offs

- [Presigned URL expiry in long-lived sessions] → After 24h the URLs in a cached API response will 403. This is the same behaviour as `image_url` today. Acceptable for a personal gallery app.
- [Bucket must be manually set to private in Cloudflare dashboard] → This change is necessary for the code change to have effect. It is an ops step outside the codebase.
