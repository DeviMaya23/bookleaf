## Why

The app has backend APIs for listing and deleting images, but the frontend has no main gallery view where users can browse and manage their photos. This is the core user-facing surface that makes the product usable.

## What Changes

- Add a root route (`/`) that displays unfoldered images by calling `GET /images?folderId=null`
- Add a folder route (`/folders/:folderId`) that displays images within a specific folder
- Display images in a responsive grid (6 columns desktop, 2 columns mobile) as thumbnail cards
- Show a loading spinner while images are being fetched
- Show an empty state message when no images exist in the current view
- Enable right-click context menu on image cards with a delete option and confirmation dialog

## Capabilities

### New Capabilities

- `fe-gallery-view`: Main gallery page — routing, image grid, folder navigation, loading/empty states, and image deletion via context menu

### Modified Capabilities

<!-- None — existing specs (image-endpoints, folder-endpoints, image-thumbnail) are not changing in requirements -->

## Impact

- Frontend: new page component(s), routing changes, API integration for `GET /images` and `DELETE /images/:id`
- Depends on existing backend endpoints: `GET /images` (with `folderId` query param) and `DELETE /images/:id`
- Depends on image thumbnail URLs being available in the image list response
