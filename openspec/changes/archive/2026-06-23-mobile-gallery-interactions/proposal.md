## Why

On mobile (touch/coarse-pointer) devices, the right panel is CSS-hidden below the `sm` breakpoint with no replacement, so users have no way to view or edit image/folder metadata. The full-resolution image viewer is reachable only via double-click, which on touch is synthesized from a double-tap — a gesture already reserved by mobile browsers for pinch-to-zoom — making entry unreliable and the viewer's desktop toolbar (rotate, flip, zoom slider, drag-pan) unsuited to touch. Both gaps need addressing without changing desktop behavior or gesture semantics in any way.

## What Changes

- Right panel gains a mobile presentation: on coarse-pointer devices, the existing `ImagePanelBody`/`FolderPanelContent` content renders inside a bottom drawer (binary open/close, backdrop dismiss) instead of being hidden. No swipe-to-dismiss or snap points; no new dependency.
- New simplified mobile image viewer (lightbox): best-effort static viewing only — fit-to-screen image, no zoom of any kind (pinch, double-tap, or slider), and no rotate/flip/drag-pan toolbar. Reuses image-loading/placeholder logic from the desktop viewer where reasonable; the desktop viewer component itself is unchanged.
- Gesture remap, gated entirely on a single capability check (`matchMedia('(pointer: coarse)')`), not viewport width, so resizing a desktop browser window narrow does not change behavior:
  - Coarse pointer: single tap on an image card opens the mobile lightbox (replacing today's "do nothing visible" single-click-opens-hidden-panel behavior).
  - Coarse pointer: the existing long-press-triggered context menu gains a "View details" item that opens the bottom drawer. This reuses the existing context-menu trigger rather than introducing a new gesture, and is gated to coarse pointer only so it doesn't add redundant menu noise on desktop (where click already opens the panel).
  - Fine pointer (mouse/desktop): `onClick` (open panel) and `onDoubleClick` (open viewer) on image cards are unchanged. The viewer is no longer reachable via synthesized double-tap on touch, which removes the existing mis-trigger bug.

## Capabilities

### New Capabilities
- `fe-image-lightbox`: Best-effort static lightbox (fit-to-screen image, no zoom) for viewing a full-resolution image on coarse-pointer devices, opened by single tap on an image card. (Note: an empty `specs/fe-image-lightbox/spec.md` already exists in the repo from a prior unfinished effort — this change populates it for the first time.)

### Modified Capabilities
- `fe-right-panel`: Adds a bottom-drawer presentation on coarse-pointer devices (reusing existing panel content components); changes the "panel SHALL NOT be rendered below the `sm` breakpoint" requirement to "panel SHALL render as a bottom drawer on coarse-pointer devices, opened via the context menu's 'View details' item rather than single tap."
- `fe-image-viewer`: On coarse-pointer devices, single tap opens `fe-image-lightbox` instead of selecting the image/opening the right panel; double-click/double-tap no longer opens any viewer on coarse-pointer devices. Fine-pointer (desktop) trigger behavior is unchanged.
- `fe-gallery-view`: Context menu gains a "View details" item, rendered only on coarse-pointer devices, that opens the right panel as a bottom drawer.

## Impact

- `frontend/src/features/right-panel/components/RightPanel.tsx` — new mobile drawer shell alongside existing desktop `<aside>` shell.
- `frontend/src/features/viewer/components/ImageViewer.tsx` — unchanged; new sibling component for the mobile lightbox.
- `frontend/src/features/gallery/components/ImageGrid.tsx` — `ImageCard`'s existing `onClick`/`onDoubleClick` wiring is unchanged; only addition is a conditional "View details" context-menu item (new `onViewDetails` prop threaded through to `AppLayout`).
- `frontend/src/app-shell/AppLayout.tsx` — the pointer-capability branch lives here: the callbacks passed as `onImageSelect`/`onImageDoubleClick` decide whether to open the lightbox or the existing panel/viewer state, instead of `ImageGrid`/`ImageCard` branching themselves. New `lightboxImage` state alongside existing `selectedImage`/`viewerImage`.
- `frontend/src/hooks/useIsCoarsePointer.ts` — new file, first cross-feature hook in the frontend (existing hooks are feature-scoped).
- `frontend/src/features/folder-sidebar/components/FolderSidebar.tsx` — referenced as the existing off-canvas drawer pattern to follow (transform), not modified.
- No new npm dependencies (no drawer/gesture library).
