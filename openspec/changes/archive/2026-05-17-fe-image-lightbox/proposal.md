## Why

Image cards in the gallery currently display only thumbnails. Users have no way to view the full-resolution version of an image. Clicking a card should open an immersive full-screen lightbox showing the high-res image fetched via a presigned URL.

## What Changes

- Left-clicking an image card opens a full-screen lightbox overlay
- The frontend calls `GET /images/:id` to retrieve the presigned R2 URL for the full-resolution image
- A spinner is shown while the URL is loading; the image is shown once ready
- The lightbox closes via the X button (top-right), clicking the backdrop, or pressing ESC
- No metadata (title, description) is shown in the lightbox — image only

## Capabilities

### New Capabilities

- `fe-image-lightbox`: Lightbox overlay triggered by clicking an image card; fetches the presigned URL on demand and displays the full-resolution image

### Modified Capabilities

<!-- None — backend `GET /images/:id` already returns `image_url` (presigned URL). No backend changes required. -->

## Impact

- Frontend only: `ImageGrid.tsx`, `src/lib/images.ts`
- Depends on existing backend endpoint: `GET /images/:id` (returns `image_url`)
- No new UI library dependencies — uses shadcn `Dialog` already present in `ImageGrid`
