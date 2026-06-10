## Context

`AppLayout` currently renders a fixed-position `FolderSidebar` (240px) with
`<main>` carrying a matching `ml-[240px]`, and conditionally renders
`RightPanel` as a flex sibling based on `selectedImage` /
(`folderPanelOpen && activeFolder`). `ImageViewer` replaces the gallery
grid inside `<main>` when `viewerImage` is set, but `<main>`'s margin and
`RightPanel` visibility are unaffected by that swap.

`ui/toggle.tsx` (a `Toggle` wrapping Base UI's toggle primitive) was added
in a recent branch but has no callers yet — this change is its first use.

## Goals / Non-Goals

**Goals:**
- One toggle, two render locations (gallery toolbar, viewer header), single
  shared boolean state.
- Hide both side panels without touching any existing selection/click logic.
- Toggling off restores panels to whatever state they'd naturally be in.

**Non-Goals:**
- Persisting focus mode across reloads/navigation.
- Animating the layout transition (can be a follow-up; plain class swap for
  now, consistent with how `RightPanel` already mounts/unmounts abruptly).
- Changing single-click/double-click semantics in any way.

## Decisions

**State location: `useState` in `AppLayout`, passed down as props.**
`AppLayout` already owns `selectedImage`, `folderPanelOpen`, and
`viewerImage` — the same component that needs to gate `FolderSidebar` and
`RightPanel` rendering. Keeping `focusMode` alongside them avoids
introducing context/global state for a single boolean used by three
siblings (`FolderSidebar` is simply omitted, `ImageViewer` and the gallery
toolbar both live directly under `AppLayout`'s JSX).

**Visibility-only toggle, no state resets.**
Per clarified requirements, turning focus mode on/off must not clear or
alter `selectedImage` / `folderPanelOpen` / `viewerImage` — it only changes
what's mounted. This means the existing `RightPanel` render condition
`(selectedImage || (folderPanelOpen && activeFolder))` gets a trailing
`&& !focusMode`, and `FolderSidebar` + the `ml-[240px]` class get wrapped in
`!focusMode` / `focusMode ? 'ml-0' : 'ml-[240px]'`. No other state
transitions change.

**Two toggle button instances, not a shared floating component.**
The button appears in two structurally different toolbars (gallery header
row vs. viewer header row) with different surrounding elements. Both are
small (`Toggle` + `Focus` icon + `aria-pressed={focusMode}` +
`onPressedChange={() => setFocusMode(v => !v)}` or equivalent), so a shared
extracted component isn't justified — duplicating ~5 lines of JSX twice is
simpler than introducing a new shared component for this. `ImageViewer`
receives `focusMode` and `onToggleFocusMode` as props since it doesn't own
the state.

**Icon and active-state styling.**
Use lucide-react's `Focus` icon. Active state styled the same way as the
mime-type filter `ToggleGroupItem`s (`aria-pressed:bg-secondary
aria-pressed:text-secondary-foreground`), so the visual language for "this
toggle is on" is consistent across the toolbar.

## Risks / Trade-offs

- **Abrupt layout shift** when toggling (sidebar/panel disappear instantly,
  `<main>` snaps to full width) → Acceptable for v1; matches existing
  `RightPanel` mount/unmount behavior, no animation precedent to follow.
- **FolderSidebar's navigation becomes unreachable while focus mode is on**
  (no folder switching, no Trash/All/Unsorted) → Intentional per proposal;
  the toggle itself is always reachable in both toolbars to exit focus mode.
- **Drag-and-drop onto folders** (`DndContext` covers `FolderSidebar`
  droppables) is moot while focus mode is on since `FolderSidebar` isn't
  rendered — no special-casing needed, `DndContext` simply has fewer
  droppable targets.
