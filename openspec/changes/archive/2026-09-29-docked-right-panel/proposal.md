## Why

The right panel mounts and unmounts conditionally based on what the user clicks, causing the gallery grid to reflow every time an image is selected, deselected, or a different folder is navigated to. This is disorienting and creates an inconsistent layout.

## What Changes

- The right panel is always present as a docked 320px column on desktop (sm and above). The grid width never changes due to panel state.
- The panel gains a collapse/expand chevron toggle in its header shell. Only the user's explicit toggle changes the layout width.
- Collapsed state is persisted to `localStorage` and survives image selection, folder navigation, and page reloads.
- A neutral content state (section title only) is shown when viewing All, Unsorted, or Trash — no image count, no actions.
- The ✕ close button is removed from all panel content modes (image, folder, selection). The chevron toggle in the shell replaces it.
- **BREAKING**: The selection panel's ✕ button (which previously exited select mode) is removed. The toolbar toggle is the sole exit affordance for select mode.
- Focus mode continues to hide the panel entirely (no change to that behavior).
- Mobile (coarse-pointer) keeps existing drawer behavior; this change is desktop-only.

## Capabilities

### New Capabilities
- `fe-right-panel-docked`: Always-present docked right panel with collapse/expand toggle, persisted collapsed state, and neutral content for All/Unsorted/Trash views.

### Modified Capabilities
- `fe-right-panel`: Several requirements change — panel is no longer conditionally mounted; the ✕ close button is replaced by the shell's chevron toggle; "Closing the selection panel exits select mode" requirement is removed; "Panel is hidden when no image is selected" requirement is replaced by the neutral content mode.
- `app-shell`: The two-panel layout description expands to a three-column layout (left sidebar + main + right panel) on desktop.

## Impact

- `frontend/src/app-shell/AppLayout.tsx` — always renders `<RightPanel>`; content is derived from state priority (selection → image → folder → neutral); `onClose` callbacks replaced by `onToggleCollapse`
- `frontend/src/features/right-panel/components/RightPanel.tsx` — new `neutral` mode; `onClose` prop removed; collapse chrome (chevron toggle) added to the `<aside>` shell; `usePersistedBoolean` hook for collapsed state
- `frontend/src/features/right-panel/components/FolderPanelContent.tsx` — ✕ button removed
- `frontend/src/features/right-panel/components/SelectionPanelBody.tsx` — ✕ button removed
- `frontend/src/app-shell/AppLayout.test.tsx` — ~12 assertions rewritten from panel presence to panel content/mode
- `frontend/src/features/right-panel/components/RightPanel.test.tsx` — `onClose` refs and `hidden sm:flex` assertions updated
