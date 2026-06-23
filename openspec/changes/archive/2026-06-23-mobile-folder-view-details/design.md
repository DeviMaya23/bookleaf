## Context

`mobile-gallery-interactions` introduced `useIsCoarsePointer()` and the `RightPanel` bottom-drawer shell, and made the *image*-card flow opt-in on coarse pointer: tapping a card opens a lightbox instead of the panel, and the panel is reached only via a new "View details" item in the image card's long-press context menu. Folder selection in the sidebar was left unchanged — `FolderSidebar.handleFolderSelect` (`FolderSidebar.tsx:73-78`) calls `onFolderSelect?.()` on every folder switch, which `AppLayout` wires unconditionally to `setFolderPanelOpen(true)` (`AppLayout.tsx:206`). This auto-opens the bottom drawer on every folder switch on mobile, with no way to opt out.

`FolderItem.tsx` already has a `ContextMenu`/`ContextMenuTrigger` (`FolderItem.tsx:70-125`) with "New subfolder"/"Rename"/"Change icon"/"Delete" — the same primitive used for image cards, which already triggers via long-press on touch with no extra wiring needed. Adding a "View details" item here follows the exact precedent of `mobile-gallery-interactions`' Decision 4 (image card's "View details" item, gated on `isCoarsePointer`).

## Goals / Non-Goals

**Goals:**
- Folder selection on coarse-pointer devices no longer forces the bottom drawer open.
- A coarse-pointer-only "View details" item on the folder's existing context menu opens the drawer explicitly, mirroring the image-card pattern.
- Zero behavior change for fine-pointer (desktop): selecting a folder continues to auto-open the sidebar panel exactly as today.

**Non-Goals:**
- No change to what the folder panel displays (`FolderPanelContent` is unchanged).
- No change to the "re-selecting the active folder is a no-op" behavior (`fe-right-panel`'s existing requirement), nor to image-card behavior at all.
- No change to whether an already-open drawer keeps updating as the active folder changes via normal navigation — that's a side effect of `RightPanel` re-rendering with the current `activeFolder`, not something this change touches.

## Decisions

### 1. Split `onFolderSelect` into selection (navigation) and explicit "view details" intent — and keep "View details" coupled to navigation, not a route-independent state

Today, `onFolderSelect` does two things conflated into one call: it navigates (handled separately, already unconditional) and it opens the panel. This change keeps `onFolderSelect` as-is for fine pointer, but on coarse pointer it becomes a no-op for the panel — only navigation happens.

**Revised during implementation** (the first pass got this wrong): the right panel's folder content is derived from `activeFolder`, which is computed from the current route (`AppLayout.tsx:97-99`), not from independent state the way `selectedImage` works for images. A first implementation had the "View details" context-menu item call a callback that only opened the panel without navigating — this opened the drawer showing whatever folder the route currently pointed at, not the long-pressed folder, which is wrong and confusing whenever the long-pressed folder isn't the active one.

The fix preserves the existing design principle (confirmed deliberate, not an oversight): on desktop, there is no UX path to view a folder's metadata without it being the active/routed folder — a single click both navigates and opens the panel, coupled together, specifically to avoid a second piece of state duplicating "which folder is the panel showing." Mobile's "View details" should reproduce that same coupling, just reachable via long-press instead of by every tap:

```
FolderSidebar.handleViewDetails(folder):
  if folder is not the active folder → navigate(`/app/folders/${folder.id}`)   [same nav handleFolderSelect already does]
  onFolderViewDetails?.()                                                      [→ AppLayout, unchanged below]
  onMobileClose?.()                                                            [closes the off-canvas sidebar drawer, mirrors handleFolderSelect]

AppLayout:
  onFolderSelect (passed to FolderSidebar, fired on every folder switch):
    fine   → setFolderPanelOpen(true); setSelectedImage(null); setAutoFocusTitle(false)   [unchanged]
    coarse → setSelectedImage(null); setAutoFocusTitle(false)                              [no panel open]

  onFolderViewDetails (new, passed to FolderSidebar, fired only from the context-menu item,
                        called after navigation has already happened so `activeFolder` is already correct):
    setFolderPanelOpen(true); setSelectedImage(null); setAutoFocusTitle(false)
```

No new state is introduced — `onFolderViewDetails` doesn't need the folder as a parameter, since by the time it runs, the route (and therefore `activeFolder`) already points at the right folder.

This mirrors `mobile-gallery-interactions`' Decision 2 (branch lives in `AppLayout`, not in the sidebar component), and Decision 4 (separate callback distinct from selection, not overloading the existing one).

Alternative considered (rejected, see above): decouple the panel's folder content from the route entirely via a new `folderPanelTarget`-style state, so "View details" could show any folder's metadata without navigating to it. Rejected because it breaks the existing, deliberate desktop invariant that the panel never shows a non-active folder's metadata, and introduces a second source of truth for "which folder is the panel showing" purely to serve a gesture that can just as easily navigate first.

Alternative considered: have `FolderItem`/`FolderSidebar` read `useIsCoarsePointer()` themselves and decide whether to call `onFolderSelect`. Rejected for the same reason `mobile-gallery-interactions` rejected it for `ImageGrid`: keeps exactly one place (`AppLayout`) deciding pointer-based behavior for panel-opening, instead of spreading that decision across components.

### 2. "View details" menu item placement: above "New subfolder", gated on `isCoarsePointer`

Same visual convention as the image card's menu (`ImageGrid.tsx:75-80`): the coarse-pointer-only item goes first, followed by a separator, then the existing items unchanged. `FolderItem` already imports `ContextMenuSeparator`; reuse it.

## Risks / Trade-offs

- **[Existing test coverage assumes unconditional auto-open]** — any `AppLayout`/`FolderSidebar` test asserting that selecting a folder opens the panel will need a fine-pointer-only framing, plus a new coarse-pointer counter-case. → Mitigated by treating this as an explicit task (mirrors `mobile-gallery-interactions` task 2.4's approach).
- **[Drawer state can go stale across folder switches]** — if a user opens the drawer via "View details" for folder A, then taps a different folder B (a plain tap, not "View details" again), the open drawer will silently start showing folder B's content (`RightPanel` re-renders with whatever `activeFolder` currently is, since `folderPanelOpen` itself isn't reset on a plain tap). This was already true before this change for the *fine-pointer* sidebar case (it never closes on folder switch either) and is not new behavior introduced here. → No mitigation planned; flagging as accepted behavior, not a regression to fix in this change. (Note: this is a different scenario from the bug found during implementation of Decision 1 above — that bug was the drawer showing the *wrong* folder immediately on open; this is the drawer staying open and updating as the user keeps navigating, which is consistent on both pointer types.)
