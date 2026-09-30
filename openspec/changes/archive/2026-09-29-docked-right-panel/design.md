## Context

The right panel is currently conditionally mounted in `AppLayout` via a three-branch ternary:
`selectedIds.size > 0` → selection | `selectedImage && !focusMode` → image | `folderPanelOpen && activeFolder && !focusMode` → folder | null.

Every transition to/from `null` causes a flex-layout reflow: the `w-80` aside appears or disappears, shrinking or expanding the main content area and reflowing the image grid. The fix is to make the panel always present in the DOM with a fixed width, and let only an explicit user action (the chevron toggle) change its width.

## Goals / Non-Goals

**Goals:**
- Panel is always mounted on desktop (sm and above); grid width never changes due to selection or navigation
- Collapse/expand chevron toggle in the panel shell; collapsed state persisted in `localStorage`
- Neutral content (view title only) for All/Unsorted/Trash — no count, no actions
- All existing content modes (image, folder, selection) continue to work as before, except the ✕ button is removed from each
- Focus mode continues to unmount the panel entirely (no layout change to that path)

**Non-Goals:**
- Mobile/coarse-pointer behavior — drawer shell unchanged
- Image count for All/Unsorted/Trash — no new backend call
- Changing the selection panel's action layout (separate follow-up)

## Decisions

### 1. Always-mounted aside, not a portal or overlay

The panel stays as a flex sibling `<aside>` in the `AppLayout` flex row. Width toggles between `w-80` (expanded) and `w-8` (collapsed — chevron strip only). This is the minimal change: no new DOM structure, no portal, no overlay. The alternative (collapsing to `w-0 overflow-hidden`) is a one-class swap and can be tried at runtime; the `w-8` strip is the implementation default since it keeps the toggle affordance visible.

### 2. Content derivation stays in AppLayout; shell owns collapse state

`AppLayout` derives a `panelContent` value using the existing priority order — selection → image → folder → neutral — and passes it to `RightPanel`. `RightPanel` owns `collapsed` state (via `usePersistedBoolean`) and renders the chevron toggle. This keeps AppLayout responsible for *what* is shown and RightPanel responsible for *how* it is shown.

`panelContent` is typed as a discriminated union:
```ts
type PanelContent =
  | { mode: 'image'; image: Image; autoFocusTitle?: boolean }
  | { mode: 'folder'; folder: Folder }
  | { mode: 'selection'; selectedCount: number; onAddToFolder: (id: string) => void; onMoveToTrash: () => void }
  | { mode: 'neutral'; viewLabel: string }
```

`RightPanel` receives `panelContent` and `onExitSelectMode` (called when needed, replaces `onClose` on the selection branch). `onClose` is removed from all modes — collapse is the panel's own concern.

### 3. `usePersistedBoolean` hook — new, small, follows `useTheme` pattern

`localStorage` is the existing precedent (`useTheme`, `ProfileMenu`). A small hook:
```ts
function usePersistedBoolean(key: string, defaultValue: boolean): [boolean, (v: boolean) => void]
```
reads on mount (defaulting if absent) and writes on change. Lives in `frontend/src/hooks/`. The key is `bookleaf-right-panel-collapsed`.

An alternative (lifting collapsed state to AppLayout) was rejected: collapsed state is purely a panel UI concern and AppLayout is already large.

### 4. Focus mode still unmounts the panel

Focus mode gates the entire `<RightPanel>` with `!focusMode` (same as today, except it now gates a single always-rendered panel rather than several branches). The selection panel exception — it survives focus mode — is preserved: when `panelContent.mode === 'selection'`, the focus mode gate is bypassed.

### 5. Chevron placement: panel shell header, not inside content components

The ✕ button today lives inside each content component (`ImagePanelBody`, `FolderPanelContent`, `SelectionPanelBody`). The new chevron lives in the `<aside>` shell in `RightPanel.tsx`, above the content area. This centralises the toggle affordance in one place and means content components no longer need any close/collapse prop. The only structural addition to the shell is a thin header row containing the chevron button.

### 6. `onClose` removal — not replaced with a selection-exit affordance in the panel

The selection panel's ✕ previously exited select mode. Under the new model, the toolbar Select mode toggle is the exit affordance. No "Done" button is added to the selection panel — the toolbar toggle is already present and functional.

## Risks / Trade-offs

- **Persistent collapsed state surprises a first-time user who collapses and forgets** → Acceptable; the chevron strip remains visible at `w-8` and is a standard affordance.
- **Tests asserting `not.toBeInTheDocument()` on the right panel break** → Expected and called out in the proposal. ~12 AppLayout tests and ~2 RightPanel tests need updating. All rewrites are mechanical (presence → mode assertion).
- **`w-8` strip may feel cramped on narrow viewports** → The panel is `hidden` below `sm` (existing class), so it never appears at narrow widths.
- **Focus mode + selection panel interaction** → Selection panel bypasses the focus-mode gate today; that exception is preserved explicitly in the derivation logic.
