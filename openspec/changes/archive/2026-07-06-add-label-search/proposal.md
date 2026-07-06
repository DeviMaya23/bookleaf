## Why

When users search by name, images with no descriptive title — but with matching AI labels — are silently excluded. Users who have opted into AI labelling should be able to opt their search into those labels too, making the search bar more useful for visually-organised libraries.

## What Changes

- Add a `search_labels` boolean query parameter to `GET /images` that, when true alongside `name`, widens the filter to also match images whose AI labels contain the search term
- Add a "Search in AI labels" toggle to the gallery filter dropdown, visible only to users with `vision_enabled = true`; when active, `search_labels=true` is sent with the name query
- No changes to the trash view, folder-view client-side search, or the Unsorted view's search behaviour

## Capabilities

### New Capabilities

- `image-label-search`: Server-side search across AI-generated image labels (`image_labels` table) as an opt-in extension to the existing title search, controlled by a `search_labels` query parameter and a per-user `vision_enabled` gate in the frontend

### Modified Capabilities

- `fe-gallery-search`: Adds the label search toggle state and passes `search_labels` to the backend when active
- `image-endpoints`: `List` repository interface and `GET /images` handler gain a `searchLabels *bool` parameter

## Impact

- **Backend**: `handler/image.go`, `usecase/image_usecase.go`, `usecase/image_pagination.go` (new `SearchLabels` field on `ListImagesParams`), `repository/image_repository.go` (`List` query gains an OR EXISTS subquery when `searchLabels=true` and `name` is non-nil)
- **Frontend**: `useGalleryControls.ts` (new `searchLabels` state, gated on `me.vision_enabled`), `GalleryToolbar.tsx` (toggle in filter dropdown), `useGalleryImages` hook or equivalent (passes `search_labels` param)
- **API contract**: additive — new optional query param on `GET /images`, no existing params changed
- **No extension changes**
