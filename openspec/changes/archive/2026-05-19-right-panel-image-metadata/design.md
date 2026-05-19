## Context

The app currently has a two-panel layout: left sidebar (240px) + main content area. The main content renders an `ImageGrid` which owns the `selectedImage` state and the lightbox. Clicking an image card opens a full-screen lightbox Dialog that shows only the image.

This change adds a third panel on the right, reroutes the image-card click to open that panel instead, and moves the full-resolution lightbox inside the panel (triggered by thumbnail click). It also changes the gallery layout from a fixed CSS grid to CSS `column-count` masonry.

On the backend, the PATCH endpoint's `updateImageRequest` struct needs `source_url` added so the panel can persist source URL edits.

## Goals / Non-Goals

**Goals:**
- Right panel displays image metadata and allows source URL editing with auto-save on blur
- Download image button fetches a presigned download URL and triggers a browser download
- Gallery becomes Pinterest-style masonry (natural aspect ratios, no card borders)
- `source_url` is patchable via `PATCH /images/:id`
- Lightbox still works, triggered from the right panel thumbnail

**Non-Goals:**
- Colour palette display (no BE support yet)
- Tags (no BE support yet)
- Editable title or notes from the panel
- Saving panel state across sessions

## Decisions

### 1. Lift `selectedImage` state to `AppLayout`

Currently `lightboxTarget` lives in `ImageGrid`. The right panel renders at the layout level (sibling to `main`), so `AppLayout` must own `selectedImage: Image | null`. `ImageGrid` receives `onImageSelect` as a prop and no longer manages its own selection state.

The lightbox state (`lightboxOpen: boolean`) stays local to `RightPanel` since it is purely a panel-internal interaction (thumbnail click).

**Alternative considered**: React Context. Rejected — the component tree is shallow; prop drilling from `AppLayout` to `ImageGrid` is one level, not worth a context.

### 2. Panel layout: push, not overlay

When the panel is open, `main` shrinks and the panel slides in as a fixed-width sibling. The masonry column count will naturally reflow. An overlay (position: fixed/absolute) would be simpler but hides content.

Panel width: `w-80` (320px), matching the design. `flex-shrink-0`.

### 3. Masonry: CSS `column-count` via Tailwind `columns-*`

The design prototype uses `column-count` with `break-inside: avoid`. Tailwind's `columns-2 md:columns-3 lg:columns-4` covers this without a JS masonry library. Cards have `mb-3 break-inside-avoid`. No fixed aspect ratio — each image card's height follows the thumbnail's natural aspect ratio.

**Alternative considered**: CSS Grid with `grid-template-rows: masonry` (subgrid masonry). Not yet supported cross-browser.

**Alternative considered**: JS masonry library (e.g. Masonry.js). Adds a dependency and requires DOM measurement; CSS `column-count` is sufficient.

### 4. Source URL: auto-save on blur

The panel has no explicit "Save metadata" button. The only editable and persistable field is source URL. It auto-saves via `PATCH /images/:id` when the input loses focus (blur event), only if the value has changed. A brief toast confirms success/failure.

**Alternative considered**: Debounced save on change. Noisier — would fire mid-typing. Blur is cleaner.

### 5. Download: FE fetches presigned URL, triggers anchor download

The download button calls `GET /images/:id/download` → receives `{ download_url }` → creates a temporary `<a href={download_url} download>` and programmatically clicks it. The BE already sets `Content-Disposition: attachment; filename=<title>.<ext>` on the presigned URL, so the filename is correct without any FE manipulation.

### 6. `source_url` added to `updateImageRequest` (BE)

`source_url *string` is added to the PATCH request struct. A `null` value clears the field; absent means unchanged. The usecase `UpdateImageParams` gains a `SourceURL **string` pointer-of-pointer following the same pattern as `FolderID`.

## Risks / Trade-offs

- [Masonry reflow on panel open] When the panel opens, `main` narrows and `column-count` reflowing causes cards to jump. → Acceptable for now; CSS transition on panel width could smooth this later.
- [Presigned download URL TTL] Download URLs expire in 5 minutes. If the user opens the panel and clicks Download after >5 min, the link will 403. → The FE fetches a fresh URL on each button click (not cached), so this is not a problem.
- [Auto-save race condition] If the user blurs the source URL input and immediately closes the panel, the PATCH may still be in flight. → Panel close does not cancel the in-flight request; PATCH is idempotent so a stale completion is harmless.

## Open Questions

- None — all decisions made during explore session.
