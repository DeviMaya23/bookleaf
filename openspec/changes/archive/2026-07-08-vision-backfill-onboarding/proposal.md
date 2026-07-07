## Why

Smart Features is now used by image label search and AI auto-categorisation, but enabling it for the first time does not process existing images — leaving users with no labels on their library until new images are uploaded. The backend backfill endpoint (`POST /me/vision/backfill`) exists but is not wired to the frontend, and users receive no disclosure that enabling this feature sends their image data to Google Vision API.

## What Changes

- Enabling the Smart Features toggle now shows a confirm modal before persisting the change
- The modal explains that image data is sent to Google Vision API and that existing images will be processed in the background
- On confirmation, the app calls `PATCH /me` (vision_enabled: true) followed by `POST /me/vision/backfill` to enqueue existing unlabelled images
- The modal appears every time the toggle is turned on (not only on first enable), reinforcing the data-processing disclosure
- Disabling the toggle retains the existing direct-call behaviour (no modal)

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `fe-vision-toggle`: Toggle-on flow changes from a direct `PATCH /me` call to a consent modal that, on confirmation, enables vision and triggers a backfill of unlabelled images

## Impact

- `frontend/src/features/settings/components/AdvancedSection.tsx` — modal state + backfill call added
- New API call wired: `POST /me/vision/backfill`
- No backend changes required; the endpoint is already implemented and registered
