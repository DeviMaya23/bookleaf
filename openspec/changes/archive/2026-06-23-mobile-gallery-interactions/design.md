## Context

`AppLayout.tsx` owns all the state this change touches: `selectedImage` (right panel), `folderPanelOpen` (right panel, folder mode), `viewerImage` (desktop image viewer), `focusMode`, and `mobileDrawerOpen` (existing sidebar drawer). `ImageGrid`/`ImageCard` are dumb — they fire `onClick`/`onDoubleClick` and call whatever callbacks `AppLayout` passed down (`onImageSelect`, `onImageDoubleClick`); they hold no device-specific logic today.

The existing mobile drawer precedent (`FolderSidebar`, `AppLayout.tsx:174-186`) has its backdrop owned by `AppLayout`, not by `FolderSidebar` itself. `RightPanel.tsx`, by contrast, is already self-contained (it branches internally on `mode: 'image' | 'folder'`). This change keeps that self-containment going rather than matching `FolderSidebar`'s split ownership — see Decision 3.

## Goals / Non-Goals

**Goals:**
- Single source of truth for "is this a touch device" so behavior never depends on viewport width.
- Zero behavior change for fine-pointer (mouse) input — same handlers, same call sites, same components.
- Reuse `ImagePanelBody`/`FolderPanelContent` unchanged inside a new mobile shell.
- No new npm dependency.

**Non-Goals:**
- Swipe-to-dismiss or snap-point gestures for the bottom drawer (binary open/close only).
- Rotate/flip/zoom-slider/drag-pan/pinch-zoom in the mobile lightbox — it's best-effort static viewing (image + close), not a parity feature with the desktop viewer.
- Swipe-between-images navigation in the lightbox (not requested).
- Changing trash, context-menu delete, or any other existing gallery behavior.

## Decisions

### 1. Pointer-capability detection: new `useIsCoarsePointer` hook

**This is a new abstraction not currently in the codebase — flagging per project convention before proceeding.**

A small hook wrapping `window.matchMedia('(pointer: coarse)')`, living at `frontend/src/hooks/useIsCoarsePointer.ts` — a new top-level directory, since existing shared hooks (`useGalleryControls`, etc.) are currently feature-scoped and this is the first cross-feature one. Returns a boolean, re-evaluated on the media query's `change` event (handles devices switching input mode, e.g. a 2-in-1 laptop).

Alternative considered: branch on viewport width (`sm` breakpoint), matching the existing `FolderSidebar`/`RightPanel` convention. Rejected per earlier discussion in this change's exploration — width and input device are different axes, and a desktop user resizing their browser narrow must keep mouse semantics (this was the explicit reason for choosing pointer-capability over viewport width).

### 2. Gesture remap lives in `AppLayout`, not `ImageGrid`/`ImageCard`

`ImageCard`'s `onClick`/`onDoubleClick` wiring (`ImageGrid.tsx:65-66`) stays completely unchanged — both events already fire identically regardless of input device. The branch happens one level up, in the callbacks `AppLayout` passes as `onImageSelect`/`onImageDoubleClick`:

```
onImageSelect:       coarse → open lightbox (new state)
                      fine   → existing: setSelectedImage + setFolderPanelOpen(false)   [unchanged]

onImageDoubleClick:  coarse → no-op
                      fine   → existing: handleImageDoubleClick                          [unchanged]
```

This is a smaller diff than originally scoped in the proposal — `ImageGrid.tsx`/`ImageCard` need no changes for the gesture remap itself, only for the new context-menu item (Decision 4). Keeping the branch at the `AppLayout` level also means there's exactly one place that reads `useIsCoarsePointer` for trigger semantics, instead of threading the flag through `ImageGrid`'s props.

**Risk:** double-tap on touch still synthesizes a `click` then a `dblclick` in most mobile browsers. Without gating, the first `click` would open the lightbox (correct) but the trailing `dblclick` would still fire `handleImageDoubleClick` and open the desktop `ImageViewer` on top of it — recreating the exact bug this change is fixing. Mitigation: `onImageDoubleClick` is a no-op when `isCoarsePointer` is true, not just "unreachable via a new gesture" — it must be an explicit guard, not an assumption that nothing will call it.

### 3. Right panel mobile shell owns its own drawer chrome (backdrop, transform), inside `RightPanel.tsx`

Rather than mirroring `FolderSidebar`'s split (backdrop owned by `AppLayout`, slide owned by the component), `RightPanel` will branch internally:

```tsx
export default function RightPanel(props: RightPanelProps) {
  const isCoarsePointer = useIsCoarsePointer()
  const content = props.mode === 'folder' ? <FolderPanelContent .../> : <ImagePanelBody .../>
  return isCoarsePointer
    ? <MobileDrawerShell onClose={props.onClose}>{content}</MobileDrawerShell>
    : <aside className="hidden sm:flex w-80 ...">{content}</aside>
}
```

Alternative considered: match `FolderSidebar` exactly, lifting backdrop state into `AppLayout`. Rejected — `RightPanel` already encapsulates its own mode branching (`image`/`folder`), so a third internal branch (shell choice) is consistent with how this component already works, and avoids growing `AppLayout`'s already-large state list (it already owns 10 pieces of state) for something that is purely presentational. `FolderSidebar`'s split predates this decision and isn't being touched.

`MobileDrawerShell` is a new small component: fixed bottom-anchored panel, `translate-y` transform + opacity-animated backdrop, mirroring `FolderSidebar.tsx:106-107`'s transform approach but on the Y axis. Closing happens via the existing `onClose` (already wired through both `ImagePanelBody`/`FolderPanelContent`'s close button) or a tap on the new backdrop.

### 4. New "View details" context menu item, gated on `isCoarsePointer`

In `ImageCard`'s `ContextMenuContent` (`ImageGrid.tsx:71-92`), add a item rendered only when `isCoarsePointer` is true, above the existing Delete/Restore items (mirrors the existing `isTrash` conditional already in that block). It needs a new callback distinct from `onSelect` (which on coarse pointer now opens the lightbox) — call it `onViewDetails`, threaded `ImageCard` → `ImageGrid` (new optional prop) → `AppLayout` (wired to the same `setSelectedImage`/`setFolderPanelOpen(false)` logic that `onImageSelect` runs today for fine pointer).

### 5. Mobile lightbox: new component, partial reuse from `ImageViewer`

New `frontend/src/features/viewer/components/ImageLightbox.tsx` (or a new `frontend/src/features/lightbox/` feature folder — naming TBD at tasks stage), populating the existing empty `specs/fe-image-lightbox/spec.md`. Triggered by the new `lightboxImage` state in `AppLayout`, rendered instead of the gallery grid the same way `viewerImage` is today (`AppLayout.tsx:206-212`), but as a sibling conditional rather than reusing `ImageViewer`'s branch — the two viewers are mutually exclusive by definition (one device gets one or the other), so no state-priority logic is needed between them.

Reused from `ImageViewer.tsx`: the blurred-placeholder/full-image load pattern (lines 98-134) and the `Esc`-to-close effect (lines 28-34) are small enough that the design defers to tasks-time judgment on extract-vs-duplicate — both are ~10-15 lines, and `ImageLightbox` not depending on `useImageTransform` (the rotate/flip/zoom/pan hook) at all means there's limited shared surface to begin with.

No zoom of any kind (pinch, double-tap, slider). The lightbox shows the image at fit-to-screen size with `object-fit: contain`, full stop — this is deliberately best-effort viewing, not a parity feature with the desktop viewer's transform tooling. This removes any gesture-library or viewport-meta question entirely.

## Risks / Trade-offs

- **[Double-tap-synthesized dblclick reopening desktop viewer]** → Mitigated by explicit no-op guard on `onImageDoubleClick` under `isCoarsePointer` (Decision 2), not an assumption.
- **[`useIsCoarsePointer` becomes the first cross-feature hook]** → Small surface (one boolean), but sets a precedent for where shared frontend hooks live; worth confirming the location before tasks.md.
- **[Bottom drawer and desktop sidebar are mutually exclusive renders inside `RightPanel`]** → If a device's pointer type changes mid-session (rare, e.g. 2-in-1 tablet undocking), the panel will re-render into the other shell on the `change` event rather than needing a remount of `AppLayout` — confirmed harmless since `RightPanel` already remounts content via `key={image.id}`/`key={folder.id}`.

