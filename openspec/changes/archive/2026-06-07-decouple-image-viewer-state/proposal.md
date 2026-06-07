## Why

The image viewer's open/visible state is currently derived entirely from `selectedImage` — the same piece of state that drives the right panel. This overload causes two bugs: navigating to a different folder leaves the viewer open showing a stale image from the old folder, and closing the right panel unexpectedly closes the viewer too (since both clear `selectedImage`).

## What Changes

- Give the image viewer its own `viewerImage` state, independent of `selectedImage` (which continues to drive the right panel only). `viewerOpen` becomes derived from `viewerImage` rather than gated on `selectedImage`.
- Closing the right panel SHALL no longer close the image viewer. The viewer's frame widens to fill the space the panel vacated (plain flex reflow — no refit/zoom recalculation).
- Navigating to a different folder SHALL dismiss both the image viewer and the right panel, returning to the gallery grid for the newly selected folder.
- Deleting the image currently open in the viewer SHALL close the viewer (in addition to the existing right-panel-deselection behavior), even if the right panel is showing a different image.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `fe-image-viewer`: introduces independent `viewerImage` state (replacing the `selectedImage`-derived gate), adds a requirement that folder navigation dismisses the viewer (and right panel), and changes the right-panel-close interaction so the viewer survives and its frame widens.
- `fe-right-panel`: corrects a stale requirement that claims clicking the thumbnail opens "the lightbox" — the lightbox was removed in favor of the image viewer (commit `8a2a3d3`) and no click-to-open wiring exists on the thumbnail today. The spec is updated to describe the thumbnail as a static display element with only the close button as an interactive affordance. This is a documentation correction, not a behavior change.

## Impact

- `frontend/src/components/AppLayout.tsx`: replace the `selectedImage`-gated `viewerOpen` with a dedicated `viewerImage` state; remove/replace the `if (!selectedImage) setViewerOpen(false)` effect; add a navigation-triggered reset effect keyed on the active folder/view; extend the image-deletion handler to also check `viewerImage`.
- No backend, API, or database changes.
