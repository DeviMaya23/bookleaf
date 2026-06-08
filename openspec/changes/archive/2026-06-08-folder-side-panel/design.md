## Context

`RightPanel` (`frontend/src/components/RightPanel.tsx`) currently has exactly one mode: it renders an `image: Image` prop and is shown/hidden by `AppLayout` based on `selectedImage` state (`AppLayout.tsx:98`, guard at line 363). It is not internally modularized — sections (thumbnail, title, notes, source URL, folders, tags, details grid, footer) are inline blocks separated by `border-b`.

Folder selection today is purely route-based navigation: `FolderSidebar.handleFolderSelect` calls `navigate(/folders/:id)` (`FolderSidebar.tsx:307-309`); there's no "active folder" component state, and folder clicks have no concept of opening a panel.

Folder updates already exist for `name` (`renameFolder`, `frontend/src/lib/folders.ts:32-39`) and `parent_id` (`moveFolder`, lines 41-49), each a thin `apiFetch` wrapper around `PUT /folders/:id`, composed with `useMutation` at the call site. There is no existing wrapper that updates `description`.

## Goals / Non-Goals

**Goals:**
- Reuse `RightPanel` as the surface for folder metadata (title, description), rather than building a new panel/modal
- Reuse the existing blur-to-save + revert-on-empty pattern already proven for image title/description (`RightPanel.tsx:202-218`)
- Keep folder selection and image selection mutually exclusive in the panel via small, symmetrical clears at each trigger site (selecting one clears the other), without inventing new precedence rules
- Keep the folder panel's displayed content perpetually correct relative to the active folder — including via navigation paths the feature doesn't directly drive (e.g. browser back/forward) — by deriving it from the URL-backed `view` rather than caching a copy

**Non-Goals:**
- Fixing the image-click layout reflow bug (separate, pre-existing issue — tracked outside this change)
- Any left-sidebar toggle or right-panel persistent/toggle mode (parked as a follow-up idea)
- Editing folder `parent_id` or other fields from the panel (only title and description, per proposal)

## Decisions

### 1. Track folder panel as a thin "open" trigger; derive its content from the URL

Add a minimal trigger to `AppLayout` — e.g. `folderPanelOpen: boolean` — set by the folder-select callback, rather than storing a copy of the selected `Folder`. The folder data actually shown in the panel (`{ id, name, description }`) is derived at render time from `view` (when `view.type === 'folder'`) plus the existing `['folders']` query, not stored separately.

**Why not store `selectedFolder: Folder | null`:** `view` comes from `useAppView()` (`AppLayout.tsx:68-76`), which is purely URL-derived via `useParams`/`useLocation` and updates on *any* URL change — including browser back/forward, which bypasses the folder-select callback entirely (the only place that calls `navigate('/folders/:id')` is `FolderSidebar.handleFolderSelect`, `FolderSidebar.tsx:311`, the same place that would set a stored `selectedFolder`). A stored copy would therefore drift out of sync with `view` whenever the URL changes through any other path — e.g. clicking folder A, then folder B, then pressing Back: `view.id` reverts to A but a stored `selectedFolder` would still hold B, showing folder B's metadata over folder A's gallery. This is the exact staleness class the `viewKey`-reset effect was introduced to prevent for `selectedImage` (see `decouple-image-viewer-state` design: "avoids an orphaned right panel showing an image from the folder just left"). Deriving the displayed folder from `view` + `folders` makes that drift structurally impossible — the panel always reflects whichever folder the URL currently points to, with no reconciliation logic needed.

**Why a trigger is still needed (not fully derived):** being on a folder's URL doesn't imply the user wants its metadata panel open — they may simply be browsing its images. The proposal's UX is explicitly gesture-driven ("clicking a folder in the sidebar... opens... the right panel"), so an explicit open/closed signal remains necessary; only the panel's *content* is derived, not whether it's showing.

**Mutual exclusivity (unchanged in shape):** selecting an image clears the folder-panel trigger (symmetrical one-line addition alongside `setSelectedImage`), and selecting a different folder clears `selectedImage`. `RightPanel` renders folder content when the trigger is on (and `view` resolves to a folder) and image content when `selectedImage` is set; the two remain mutually exclusive by construction.

**Alternative considered:** discriminated union (`{ type: 'image' | 'folder', ... } | null`). Rejected for the same reason as before — larger refactor than the feature warrants, with no other panel content types on the horizon.

### 2. New click handling: differentiate "switch folder" from "reselect active folder"

`FolderSidebar` already derives `isActive` from the route (`view.type === 'folder' && view.id === folder.id`, line 91). The folder click handler (`handleFolderSelect` / `onSelect`, lines 307-309, 133) gains a new responsibility: notify `AppLayout` of the clicked folder via a new callback prop (e.g. `onFolderSelect`), but only when the clicked folder differs from the currently active one — when it's the same folder, the callback is skipped entirely, so panel state (whatever it currently shows) is left untouched. This keeps the no-op behavior explicit at the trigger site rather than requiring `AppLayout` to diff folder IDs.

**Why here, not in `AppLayout`:** `FolderSidebar` already computes `isActive` for the highlight; reusing that comparison at the click site avoids duplicating "is this the active folder" logic in two places.

### 3. New `FolderPanelContent` sub-component, `RightPanel` becomes a thin mode switch

Introduce `frontend/src/components/FolderPanelContent.tsx` containing the full folder-specific body — including its own close button (placed in the title-bar corner, since folder mode has no thumbnail to overlay it on) and no footer (folders have no download action). `RightPanel` keeps only the outer 320px chrome (width, border, scroll container) and conditionally renders either the existing image body (extracted as `ImagePanelBody`, retaining its thumbnail-overlaid close button and footer) or `<FolderPanelContent folder={...} onClose={...} />`.

**Why:** `RightPanel` is already a large, non-modularized component; adding a second full content mode inline would make it harder to follow and test. A sibling component mirrors the existing precedent of extracting reusable panel pieces (e.g. `FolderInput`). It also gives the new `fe-folder-panel` capability a concrete, independently-testable unit, matching how the proposal scopes it as its own spec.

**Alternative considered:** branching inline inside `RightPanel`. Rejected — would roughly double the size of an already-large component for a mode that shares only the outer chrome, not the body structure.

### 4. New `updateFolderDetails` lib function for `{ name, description }`

Add `updateFolderDetails(getToken, id, { name?, description? })` to `frontend/src/lib/folders.ts`, following the existing one-function-per-use-case convention (`renameFolder` for name-only renames via the dialog, `moveFolder` for reparenting) — all three hit `PUT /folders/:id` with different body shapes.

**Why not extend `renameFolder`:** it's used by `FolderNameDialog` for the rename flow, which has no concept of description. Overloading it with an optional `description` param would blur its single-purpose name and force unrelated call sites to reason about a field they don't use.

### 5. Reuse the blur-to-save pattern verbatim

`FolderPanelContent` mirrors `RightPanel`'s existing title/description handling: local `useState` mirroring the prop, `useRef` holding the original value (reset on `folder.id` change), `onBlur` diff-and-revert-if-empty for title, diff-and-allow-empty (saved as `null`) for description, both funneling through one `useMutation` wrapping `updateFolderDetails` that invalidates `['folders']` and toasts on success/error.

**Why:** This is the exact validation/persistence shape the proposal calls for (title required non-empty with revert, description optional and clearable), already implemented and proven for images (`RightPanel.tsx:202-218`, `85-95`). Reimplementing it differently for folders would be inconsistent for no benefit.

## Risks / Trade-offs

- **[Risk]** Two independent pieces of state (`selectedImage`, the folder-panel open trigger) could theoretically both end up active if a future code path sets one without clearing the other, causing ambiguous panel content → **Mitigation**: enforce clearing the other at every setter site (small, localized change); if a third content type is ever added, revisit decision #1 and consolidate into a union then.
- **[Risk]** Deriving folder content from `view` + the `['folders']` query means the panel briefly has no content to show if `folders` hasn't loaded yet (e.g. cold cache) even though the trigger is on → **Mitigation**: `AppLayout` already fetches `['folders']` with a 60s `staleTime` for the sidebar itself, so by the time a folder can be clicked, the data is already resident; a momentary `undefined` lookup can be treated as "nothing to render yet", matching how `RightPanel` already tolerates `allFolders` being briefly undefined.
- **[Risk]** `FolderSidebar` gaining a new "notify parent of folder selection" responsibility on top of its existing route-navigation duties slightly increases its surface area → **Mitigation**: the new callback is a single optional prop following the same shape as existing `onSelect`-style callbacks already passed through `FolderItem`.
- **[Trade-off]** Adding a third folder-update wrapper (`updateFolderDetails`) alongside `renameFolder`/`moveFolder` means three functions hit the same endpoint with overlapping body shapes → accepted, since it matches the existing convention and keeps each call site's intent clear; consolidating them would be a separate refactor outside this change's scope.
