## Context

The Smart Features toggle in `AdvancedSection` currently calls `PATCH /me` directly on every toggle change. The backend backfill endpoint (`POST /me/vision/backfill`) is fully implemented and registered but has no frontend caller. Users enabling Smart Features for the first time (or after a period with it disabled) have no way to process their existing unlabelled images.

## Goals / Non-Goals

**Goals:**
- Intercept the toggle-on action with a consent modal explaining the Google Vision API data processing
- On confirmation, enable vision and trigger backfill in sequence
- Keep the disable path unchanged

**Non-Goals:**
- Showing a count of images to be processed (requires an extra round trip, not worth the cost)
- Tracking "first enable only" — the modal shows on every enable (consent is valid each time)
- Any backend changes — the endpoint is already implemented

## Decisions

### Decision: Separate `VisionBackfillConfirmDialog` component

The confirm modal lives in a new `VisionBackfillConfirmDialog.tsx` in `features/settings/components/`. It is composed as a separate component for the same reason as `DeleteFolderDialog` — inlining the dialog JSX would make `AdvancedSection` harder to read.

**Alternative considered:** Inline modal state directly in `AdvancedSection`. Rejected — same reasoning as above.

**Implementation note:** The component does not use the `DialogContent` composite component. Because `VisionBackfillConfirmDialog` is rendered inside the settings modal (itself a Base UI dialog), Base UI treats it as a nested dialog and suppresses its backdrop by default. To work around this, the component composes `DialogPortal + DialogPrimitive.Popup` directly, passing `forceRender` to `DialogOverlay` (forces the backdrop to render for nested dialogs) and `container={document.body}` to `DialogPortal` (ensures the portal lands outside the settings modal's portal subtree so stacking and `backdrop-filter` work correctly). `onClick={onCancel}` on `DialogOverlay` handles click-outside dismissal, which Base UI's pointer-dismissal mechanism does not fire reliably in this nested context.

### Decision: `backfillVisionLabels` added to `features/auth/lib/me.ts`

The backfill endpoint is `POST /me/vision/backfill`, which is semantically part of the `/me` resource. Adding a `backfillVisionLabels(getToken)` function to `me.ts` keeps all `/me` API calls in one place.

### Decision: Sequential calls, non-atomic failure handling

On modal confirm, two calls fire in sequence:
1. `PATCH /me { vision_enabled: true }` — if this fails, show error toast, close modal, toggle stays off
2. `POST /me/vision/backfill` — runs only after step 1 succeeds. If this fails, vision is **already enabled** (step 1 committed) — we do not roll it back. A warning toast is shown ("Smart Features enabled, but existing images could not be queued for processing. Try again later."). Vision will work for newly uploaded images regardless.

**Alternative considered:** A single combined endpoint. Rejected — unnecessary backend change for a purely FE concern.

### Decision: Toggle state during modal

The toggle is controlled: `checked` reflects the `me` query cache (`vision_enabled`). When the modal is open (user clicked to enable), the toggle visually stays in its current off state — the change is not applied until confirmed. The switch is not disabled during the modal; if the user closes the modal, the toggle is simply unchanged.

## Risks / Trade-offs

- **Backfill silently no-ops if vision_enabled is false at job-processing time** — this is existing behaviour in `ProcessVisionLabelling`. Since we enable vision before calling backfill, the race condition window (PATCH succeeds, backfill fires, job runs before enable propagates) is negligible and the job would simply re-run on next trigger or upload.
- **User re-enables frequently** — the modal appears every time, which could feel repetitive. Accepted trade-off: the disclosure is accurate each time (any new images added while disabled will now also be queued).
