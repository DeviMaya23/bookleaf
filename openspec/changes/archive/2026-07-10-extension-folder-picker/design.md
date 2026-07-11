## Context

The extension popup currently has two views managed by a single `view` state in `App.tsx`: `"main"` and `"settings"`. The Settings view renders a native `<select>` for default folder selection, which cannot express the folder hierarchy the user has configured in the web app.

`GET /folders` already returns `parent_id: string | null` on each folder. The FE's folder sidebar (`frontend/src/features/folder-sidebar/lib/folderTree.ts`) provides two battle-tested pure functions — `buildFolderTree` and `filterFolderTree` — that have no external dependencies and can be ported directly.

## Goals / Non-Goals

**Goals:**
- Replace the `<select>` with a dedicated FolderPicker panel that renders the folder hierarchy with indentation.
- Filter input narrows the tree using the same logic as the FE (parent always shown when a descendant matches; non-matching children hidden).
- Confirm-on-click: selecting a row immediately persists and returns to Settings.
- Settings row becomes a button showing the current folder name (or "None" when unset).

**Non-Goals:**
- Folder creation or rename from within the picker.
- Collapsible/expandable tree nodes (all levels always rendered).
- Persistent filter state across panel opens.
- Per-save folder override (separate future feature).

## Decisions

**D1: Extend the view stack to `"main" | "settings" | "folder-picker"`**

The existing `view` state in `App.tsx` is the natural place to add a third entry. Navigation flows linearly: main → settings → folder-picker, with back buttons at each level. Passing `onOpenFolderPicker` as a prop from `App` to `Settings`, and `onBack` / `onSelect` from `App` to `FolderPicker`, keeps each component unaware of the broader navigation context.

*Alternative considered:* Render the picker as an overlay within Settings. Rejected — the popup is only 320px wide; an overlay clips at the bottom and leaves the Settings rows visible behind it, which is distracting and adds unnecessary z-index complexity.

**D2: Port `buildFolderTree` and `filterFolderTree` into `extensions/src/lib/folderTree.ts`**

The FE utilities are pure functions with no imports beyond a local `Folder` type. Copying them avoids a cross-package import boundary (the extension build cannot import from `frontend/`) and keeps the logic identical. The extension defines its own `FolderNode` type locally (matching the FE shape: `{ id, name, parent_id, children }`).

*Alternative considered:* Extract into a shared workspace package. Overkill for two small functions; adds build complexity.

**D3: `getFolders()` return type gains `parent_id: string | null`**

The backend already returns this field. The only change is making the TypeScript type reflect reality. No API call changes are needed.

**D4: Filter applied only when the query is non-empty; full tree shown otherwise**

When the filter input is blank, `filterFolderTree` is skipped and the full tree is rendered. This avoids the cost of calling the filter on every render with an empty string, and matches the FE's pattern (`trimmedFolderFilter ? filterFolderTree(...) : tree`).

**D5: FolderPicker fetches folders on mount, same as Settings today**

`getFolders()` is called inside a `useEffect` when `FolderPicker` mounts. This keeps the data fresh without adding a shared fetch layer. The select is disabled while loading (same degradation pattern as the old `<select>`).

*Alternative considered:* Pass the folder list as a prop from Settings (so only one fetch occurs per Settings open). Rejected — it couples the components unnecessarily and requires Settings to own state it doesn't use for rendering.

## Risks / Trade-offs

**Stale tree on re-open** — The folder list is fetched fresh each time FolderPicker mounts, so it is always current. If the user adds a folder in the web app while Settings is open, they'd need to navigate away and back to see it. → Acceptable; same limitation as the old `<select>`.

**Deep trees and scroll** — The popup can scroll vertically (no hard max-height is set by the browser on MV3 popups up to ~600px). A very deep or wide tree may require scrolling. No pagination needed for typical use. → Acceptable.

**`visited` guard in `buildFolderTree`** — The ported function uses a `visited` set to protect against duplicate child insertion if the API returns malformed data (a node appearing twice). This guard is preserved from the FE unchanged.
