## Why

Users need a way to download the original high-resolution version of their images directly from the app. Currently there is no download endpoint, so users have no self-service path to retrieve their full-quality files.

## What Changes

- Add a new `GET /images/:id/download` endpoint that returns a short-lived presigned R2 URL for the original (high-res) image object, triggering a browser file download.

## Capabilities

### New Capabilities

- `image-download`: Endpoint that generates and returns a presigned download URL for the original high-resolution image, scoped to the authenticated user.

### Modified Capabilities

<!-- No existing spec-level requirement changes. -->

## Impact

- **API**: New `GET /images/:id/download` route on the image router
- **Storage**: Uses existing R2 client to generate a presigned GET URL with a `Content-Disposition: attachment` header to force download
- **Auth**: Endpoint is authenticated; user must own or have access to the image
- **Code**: Touches image handler, image service, and potentially image repository layers
