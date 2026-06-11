## Context

`frontend/src/` is flat: `components/` (29 files), `lib/` (12 files),
`hooks/` (2 files), `pages/` (4 files). Files for unrelated UI areas — the
gallery grid, the folder sidebar, the image viewer, the right-hand metadata
panel, upload modals, auth — all sit side by side with no grouping.

Three files have grown disproportionately by absorbing responsibilities
beyond their nominal job:

- **`AppLayout.tsx` (~650 lines)**: in addition to being the app shell
  (sidebar | main | right panel composition + `DndContext`), it directly
  owns gallery search/sort/filter state and the entire filter/sort dropdown
  JSX (~110 lines), image-selection state for the right panel (~50 lines),
  viewer/focus-mode state (~15 lines), and file-drop-to-upload handling
  (~90 lines). Only ~280 of its lines are actually shell composition.
- **`FolderSidebar.tsx` (~495 lines)**: bundles pure tree-building utilities
  (`buildFolderTree`/`filterFolderTree`, ~40 lines, zero React), three
  cohesive sub-components (`FolderItem`, `UnsortedEntry`, `RootDropZone`,
  ~150 lines), folder CRUD mutations (~40 lines), and five separate
  dialog-state/JSX pairs (~250 lines combined with the nav list) — three of
  which (new root folder, new subfolder, rename) are the same
  `FolderNameDialog` triggered from three places.
- **`RightPanel.tsx` (~387 lines)**: `ImagePanelBody` mixes tag
  resolve-or-create business logic (~35 lines), three near-identical
  blur/dirty-check field handlers for title/description/source URL
  (~30 lines of repetition), and three queries plus a sync effect for
  image/folders/tags (~25 lines), alongside the actual form JSX.

The shared domain layer (`lib/images.ts`, `lib/folders.ts`, `lib/tags.ts`,
`lib/thumbnail.ts`, `lib/view.ts`) is imported across nearly every component
listed above — these are cross-cutting domain types/APIs, not feature-local.

## Goals / Non-Goals

**Goals:**
- Establish a `features/<feature>/` convention (plus `app-shell/`) that
  future agent-driven changes follow by default.
- Reduce the responsibility scope of `AppLayout`, `FolderSidebar`, and
  `RightPanel` via the specific extractions identified in exploration —
  not just relocate them.
- Preserve all existing behavior exactly; this is a zero-functional-change
  refactor verified by the existing (relocated/split) test suite.

**Non-Goals:**
- No new features, UI changes, or behavior changes of any kind.
- No new external dependencies.
- Not eliminating cross-cutting state from `app-shell` entirely —
  drag-and-drop orchestration and view-routing are inherently cross-feature
  and remain there by design (see Decision D4).
- No backend or browser-extension changes.

## Decisions

### D1: Feature naming avoids collision with `lib/` domain modules

Feature folders are named `gallery`, `viewer`, `folder-sidebar`,
`right-panel`, `upload`, `auth` (plus `app-shell`) — **not** `folders` or
`image-details`. `lib/folders.ts`, `lib/images.ts`, and `lib/tags.ts` remain
the shared domain layer (D2); naming the UI features `folder-sidebar` and
`right-panel` keeps "where is the folder *domain* code" (`lib/folders.ts`)
and "where is the folder *sidebar* UI" (`features/folder-sidebar/`)
unambiguous to grep. This also aligns with the existing spec names
(`fe-folder-panel`, `fe-right-panel`).

### D2: Domain layer (`lib/`) is unchanged

`images.ts`, `folders.ts`, `tags.ts`, `thumbnail.ts`, `view.ts`, `utils.ts`,
`browser.ts`, `fracdex.ts`, `api.ts` stay in `lib/` as-is. Each is consumed
by 4+ features; moving any of them under a single feature would force the
others to reach across feature boundaries, defeating the purpose of the
split.

The only addition to this shared layer is `resolveOrCreateTags` →
`lib/tags.ts`, since tag resolve-or-create is general-purpose tag-domain
logic. By contrast, `buildFolderTree`/`filterFolderTree` are
folder-sidebar-specific (only `FolderSidebar` uses them) and go to
`features/folder-sidebar/lib/folderTree.ts`, not `lib/`.

### D3: `FolderInput` stays in `right-panel`

`FolderInput`/`FolderInput.test.tsx` are only consumed by `RightPanel` today.
They move to `features/right-panel/components/`. If `folder-sidebar` ever
needs a folder-picker, promote it to shared `components/` at that point
(YAGNI) — premature promotion is the kind of speculative abstraction
CLAUDE.md's Decision Boundaries asks us to avoid.

### D4: Sequencing — self-contained features first, `app-shell` last

Order of work (also reflected in `tasks.md`):

1. **`folder-sidebar`** and **`right-panel`** — fully self-contained.
   Each gets its directory + internal extractions (tree utils/FolderItem/
   UnsortedEntry/RootDropZone/useFolderMutations/dialog consolidation for
   folder-sidebar; `useFieldAutosave`/`useImageDetailsData`/
   `resolveOrCreateTags` for right-panel) without touching `AppLayout`.
2. **`viewer`**, **`upload`**, **`auth`** — pure moves, no internal
   extraction needed (these files are already cohesive).
3. **`gallery`** — move `ImageGrid`/`MasonryLayout`, then extract
   `useGalleryControls` + `GalleryToolbar` *out of* `AppLayout` into this
   feature. This is the first step that shrinks `AppLayout`.
4. **`app-shell`** — move `AppLayout` into `app-shell/`, extract
   `useAppView` and `useAppDragAndDrop`, and wire in `GalleryToolbar` from
   step 3. This is the final integration step and the only one that touches
   nearly every other feature's public exports.
5. **`CONVENTIONS.md`** update, documenting the resulting structure.

Steps 1–3 can each be verified independently (build + tests green) before
moving on. Step 4 is the highest-risk step and should be done last, once
everything it depends on already exists in its target location.

### D5: Test strategy — split alongside extraction, keep one integration test

For each extracted hook/component, a new focused test file is created in the
same step (e.g. `useFolderMutations.test.ts`, `useFieldAutosave.test.ts`,
`GalleryToolbar.test.tsx`). The original component's test file is pruned to
cover only what remains in that component.

`AppLayout.test.tsx` (relocated to `app-shell/AppLayout.test.tsx`) keeps
**at least one integration-level test** for cross-feature interactions that
no single unit owns — e.g. double-clicking an image sets both
`selectedImage` (right-panel) and `viewerImage` (viewer); deleting the
selected/viewed image clears both. These flows are the reason this state
lives in `app-shell` at all (D4 step 4), so shell-level tests must keep
covering them even after the per-feature unit tests move out.

### D6: `FolderSidebar` dialog consolidation

The three `FolderNameDialog` instances (new root folder, new subfolder,
rename) collapse into one piece of state:

```ts
type NameDialogState =
  | { mode: 'create-root' }
  | { mode: 'create-sub'; parent: Folder }
  | { mode: 'rename'; target: Folder }
  | null
```

...and one `<FolderNameDialog>` render whose `title`, `initialValue`, and
`onSubmit` are derived from `NameDialogState`. The delete-confirmation and
empty-trash dialogs are unaffected (different shape: confirm, not
name-input).

### D7: `useFieldAutosave` hook shape

```ts
function useFieldAutosave<T>(
  value: T,
  onSave: (value: T) => void,
  options?: { isEmpty?: (value: T) => boolean }
): { value: T; onChange: (value: T) => void; onBlur: () => void }
```

Replaces the three `useState` + `useRef(orig)` + `useEffect` (reset on
image change) + `onBlur` (compare-and-save, with the title field's
"don't save empty" special case via `options.isEmpty`) groups in
`ImagePanelBody`. Lives in `features/right-panel/hooks/useFieldAutosave.ts`
— it's specific to this autosave-on-blur pattern used three times in one
component, not promoted to shared `hooks/` (no other consumer today).

## Risks / Trade-offs

- **Large diff surface across steps 3–4** (gallery + app-shell) — these are
  the only steps that modify `AppLayout.tsx` itself, and they're sequenced
  last specifically to minimize the window where `AppLayout` is
  simultaneously being restructured and depended on by newly-moved features.
- **Merge conflict risk with concurrent FE branches** touching `AppLayout`,
  `FolderSidebar`, or `RightPanel` — recommend landing this refactor when no
  other frontend feature branches are in flight, or rebasing those branches
  after this merges.
- **`useFieldAutosave` and `useImageDetailsData` are new hook abstractions**
  scoped to `features/right-panel/`. Flagging per CLAUDE.md Decision
  Boundaries: both are single-feature, single-file-origin extractions (not
  new architectural layers), so no further confirmation gate is proposed
  beyond this design doc.
- **Splitting `AppLayout.test.tsx`** could silently drop coverage if an
  extracted unit's tests don't fully replicate what the original test
  exercised — mitigated by D5 (write new tests in the same step as
  extraction, prune originals only after).
