## Context

`AppLayout.tsx` currently overloads `selectedImage` to drive three things at once: the right panel's content, the image viewer's content (`<ImageViewer image={selectedImage} ... />`), and the viewer's open/closed gate (via `useEffect(() => { if (!selectedImage) setViewerOpen(false) }, [selectedImage])`). Because RightPanel's close button also nulls `selectedImage`, closing the panel collaterally closes the viewer. Because folder navigation is route-driven and never touches `selectedImage`/`viewerOpen`, the viewer survives navigation showing a stale image.

## Goals / Non-Goals

**Goals:**
- Give the viewer an independent source of truth for which image it shows and whether it's open.
- Make the right panel's open/closed state fully independent of the viewer's.
- Dismiss both the viewer and the right panel when the active folder changes.
- Close the viewer when its currently-displayed image is deleted, regardless of right-panel state.

**Non-Goals:**
- No refit/zoom recalculation when the viewer's frame resizes (e.g. on right-panel close) — the frame simply widens via existing flex layout, image stays at current zoom (per proposal: "plain flex reflow — no refit/zoom recalculation").
- No change to how the right panel shows folder metadata, or to single-click/double-click selection semantics beyond what's listed in the proposal.
- No change to `ImageViewer`'s internal transform behavior (`fe-image-viewer-interactions` is untouched).

## Decisions

### 1. Replace `viewerOpen: boolean` + shared `selectedImage` with a single `viewerImage: Image | null`

`AppLayout` will hold `viewerImage` as its own `useState<Image | null>`, separate from `selectedImage`. The viewer renders when `viewerImage !== null`; `<ImageViewer image={viewerImage} onClose={() => setViewerImage(null)} />`.

This collapses "is the viewer open" and "what does the viewer show" into one piece of state (mirroring how `selectedImage` already does double duty for the right panel) — there's no scenario where the viewer is "open" without an image, so a separate boolean is redundant once the image itself is independent state.

**Alternative considered**: keep `viewerOpen` + `selectedImage`, and special-case RightPanel's `onClose` to skip nulling `selectedImage` while the viewer is open. Rejected — this is the "patch the gate" approach discussed in exploration; it requires the right panel to know about the viewer's state to decide how to behave, which is the same kind of cross-concern coupling causing the current bug, just relocated.

### 2. Double-click sets `viewerImage` directly; it no longer depends on `selectedImage`

`handleImageDoubleClick` currently does `setSelectedImage(img); setViewerOpen(true)`. It becomes `setSelectedImage(img); setViewerImage(img)` — both set independently from the same source event. The right panel and viewer end up showing the same image at open time (matching current/expected UX — see `fe-image-viewer` Scenario "Double-clicking a card opens the viewer"), but each can subsequently change independently (e.g. right panel closes without affecting the viewer).

### 3. Remove the `if (!selectedImage) setViewerOpen(false)` effect; add explicit resets where needed

This blanket effect is the root coupling. It's removed outright. Its job is replaced by explicit, scoped handling:
- **Right panel close** (`onClose={() => setSelectedImage(null)}`): unchanged — only affects `selectedImage`/right panel. Viewer is untouched, so it stays open and its frame widens via the existing flex layout once `<aside>` unmounts.
- **Image deletion** (`onImageDeleted`): extend the existing check from `if (selectedImage?.id === id)` to also check `if (viewerImage?.id === id) setViewerImage(null)`, alongside the existing `selectedImage` reset. The two checks are independent — either, both, or neither may fire depending on what's open.
- **Folder navigation**: new effect (see Decision 4).

### 4. New effect resets both `viewerImage` and `selectedImage` on folder/view change

A `useEffect` keyed on the active view identifier (`folderId` / `view`, derived from `useAppView()`) sets `setViewerImage(null)` and `setSelectedImage(null)` (and `setAutoFocusTitle(false)` to match the existing reset pattern at the RightPanel close site). This dismisses both the viewer and the right panel — a clean slate on navigation, per the proposal's "Option 2" decision (avoids an orphaned right panel showing an image from the folder just left).

This mirrors the existing pattern of scoped `useEffect`s in `AppLayout` reacting to state/prop changes (e.g. the very effect being removed), rather than introducing a new state-management approach.

## Risks / Trade-offs

- **[Risk]** The reset effect fires on initial mount (since the dependency value is defined from the start), which would needlessly call the setters when there's nothing to reset. → **Mitigation**: this is harmless — `setViewerImage(null)` on an already-`null` state is a no-op re-render skip in React; no special-casing needed (consistent with how the existing `selectedImage` effect behaves on mount).
- **[Risk]** Splitting `selectedImage` and `viewerImage` means a future feature that needs "the image currently being looked at, however it's being looked at" must explicitly reconcile both. → **Mitigation**: no such feature exists today; if one arises, it can derive a combined value (`viewerImage ?? selectedImage` or similar) at the call site rather than re-coupling the underlying state.

## Migration Plan

Pure frontend state refactor within `AppLayout.tsx` — no data migration, no API changes, no feature flag needed. Ships as a single FE change; rollback is a plain revert.

## Open Questions

None — exploration resolved the open questions (dismiss-always on navigation; right-panel close leaves viewer open with widened frame; split state over patching the gate).
