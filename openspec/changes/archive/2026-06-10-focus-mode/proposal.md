## Why

Users who want to browse their gallery or inspect a single image without
distraction currently have no way to dismiss the persistent left sidebar.
The right panel can already be closed, but the left sidebar is always
present, eating 240px of width even when the user just wants maximum space
for images. A single "focus mode" toggle lets users hide both panels at
once and reclaim the full viewport for the gallery or image viewer.

## What Changes

- Add a "Focus" toggle button (using the existing `Toggle` component from
  `ui/toggle.tsx`) to the leftmost position of the top toolbar in:
  - the gallery view (left of the search input)
  - the image viewer header (left of the Back button)
- When focus mode is active:
  - `FolderSidebar` is not rendered, and `<main>` expands to full width
    (loses its `ml-[240px]` margin)
  - `RightPanel` is not rendered, regardless of `selectedImage` or
    `folderPanelOpen` state
- Existing click behavior (single-click sets `selectedImage`, double-click
  opens `ImageViewer`, folder selection sets `folderPanelOpen`) continues
  unchanged underneath — focus mode is purely a render-visibility switch
  layered on top of existing state. Toggling focus mode off immediately
  reflects whatever selection state has accumulated while it was on.
- Focus mode state is plain session-only React state (`useState` in
  `AppLayout`) — no persistence across reloads or navigation.
- `ImageViewer` gains two new props (`focusMode`, `onToggleFocusMode`) to
  render its own copy of the toggle button.

## Capabilities

### New Capabilities
- `focus-mode`: a toggleable UI state that hides the left sidebar and right
  panel to give the gallery/image viewer the full viewport, accessible via
  a toggle button in the gallery and image viewer toolbars.

### Modified Capabilities
- `app-shell`: the "Two-panel application shell" requirement (sidebar
  always visible) gains an exception — the sidebar is not rendered while
  focus mode is active.
- `fe-right-panel`: the "Right panel opens when an image card is clicked"
  requirement gains an exception — the right panel is not rendered while
  focus mode is active, even if an image is selected or a folder panel
  would otherwise be open.

## Impact

- `frontend/src/components/AppLayout.tsx`: new `focusMode` state, toggle
  button in the gallery toolbar, conditional rendering of `FolderSidebar`
  and `RightPanel`, conditional `<main>` margin class.
- `frontend/src/components/ImageViewer.tsx`: new `focusMode` /
  `onToggleFocusMode` props, toggle button in the viewer header.
- `frontend/src/components/ui/toggle.tsx`: existing component, no changes
  expected — first real usage.
- No backend, API, or extension changes.
