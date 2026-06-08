## Context

The backend contract already exists (`image-list-sort`, archived): `GET /images` accepts optional `sort` (`created_at` | `title`) and `direction` (`asc` | `desc`), defaults to `created_at DESC` for non-folder views and `position ASC` for folder views when omitted, and the cursor stays opaque to the client. This phase only has to *consume* that contract from the frontend — no new backend capability is needed.

Today's frontend toolbar (`AppLayout.tsx:234-269`) has a search input and an upload split-button; there is no sort/filter affordance. `ImageGrid.tsx` already threads `view` and `searchTerm`/`debouncedSearchTerm` down from `AppLayout`, builds its query key/fetcher per view (`queryKeyFor`/`fetcherFor`, lines 103-121), and gates folder-only behaviors (`isFolderView`, drag, `SortableContext`, the `sortEndTrigger` effect) on `view.type === 'folder'` in three places (lines 53, 152, 237, plus the `SortableContext` wrap at 327-330).

The design handoff ("Filter & Sort Options.html", Option A) shows a standalone sort icon button opening a panel with a radio list (`SORT_FIELDS`) and a direction row whose label depends on the selected field (`DIR_LABELS`). The handoff is a static prototype with its own bespoke `Panel`/`RadioDot` components — it is not bound to this app's component library, so the actual implementation should map its visual intent onto primitives that already exist here.

The codebase already has `DropdownMenuRadioGroup`/`DropdownMenuRadioItem` (`src/components/ui/dropdown-menu.tsx:182,191`), used elsewhere via Base UI (per project convention — not Radix, no `asChild`). The existing upload split-button (`AppLayout.tsx:253-267`) is the closest precedent for "icon button that opens a small menu in the toolbar."

## Goals / Non-Goals

**Goals:**
- Add a sort control to the gallery toolbar that lets users choose sort field + direction, scoped to exactly `created_at`/`title` (matching the backend allow-list).
- Make `Manual` a first-class, top-of-list option in folder views that maps to "send no `sort`/`direction` params" — reproducing today's `position ASC` behavior exactly, not a new backend mode.
- Reset sort to the view's default whenever the user switches views, mirroring the existing search-term reset (`AppLayout.tsx:93`).
- Disable drag-to-reorder whenever the active sort isn't `Manual`, extending the existing `isFolderView`/`view.type === 'folder'` gates rather than introducing a new gating mechanism.
- Reuse existing UI primitives (`DropdownMenu` + `DropdownMenuRadioGroup`/`RadioItem`, `buttonVariants`) for the sort control rather than building bespoke panel/radio components from scratch.

**Non-Goals:**
- `file_size`/`dimensions` sort fields — not in the backend allow-list; deferred until backend support exists.
- Persisting sort choice across sessions or views — sort is session-local and per-view, reset on navigation (same lifecycle as search).
- Options B/C from the design handoff (combined "Sort & filter" button, inline chip bar) — only Option A's two-icon-button layout is in scope.
- Filter functionality (tags/file type/folder) shown alongside sort in the handoff — that's a separate capability, not part of this change.

## Decisions

### 1. `Manual` is represented as `sortBy: 'manual'` meaning "omit `sort`/`direction` from the request"

The FE sort state is `'manual' | 'created_at' | 'title'` (plus a direction for the latter two). When `sortBy === 'manual'`, `getImages` is called with no `sort`/`direction` params at all — the exact same call shape as today. This keeps the FE↔BE contract honest: `Manual` isn't a value the backend knows about, it's the FE's name for "don't ask for a particular order, give me what the view defaults to" (which, for folders, is `position ASC`).

**Alternative considered**: send an explicit `sort=position` or `sort=manual` to the backend. Rejected — the backend allow-list doesn't include it (would be a `400`), and adding it would mean extending the backend contract for something that's purely a frontend framing of "no sort requested."

### 2. Sort state lives in `AppLayout`, passed down to `ImageGrid` as props — same lifecycle as `searchTerm`

`AppLayout` already owns `searchTerm`/`debouncedSearchTerm` and resets them on `viewKey` change (`AppLayout.tsx:80-81, 93`). Sort state (`sortBy`, `sortDir`) follows the identical pattern: declared in `AppLayout`, reset to the view's default in the same `useEffect` keyed on `viewKey`, and passed to `ImageGrid` as props. `ImageGrid` only needs `sortBy`/`sortDir` to (a) build its query key/fetcher params and (b) compute the drag-disable condition — it doesn't need to own the selection UI.

**Alternative considered**: keep sort state inside `ImageGrid` (closer to where it's consumed). Rejected — the sort *control* (button + panel) renders in `AppLayout`'s toolbar row, alongside the search input and upload button, mirroring where the design handoff places it. Splitting "owns the value" from "renders the control" across the same component pair that already does this for search avoids introducing a second state-sharing pattern.

### 3. Per-view defaults: folder views default to `Manual`; All/Unsorted/Trash default to `Date added` / `desc`

This makes the new control *describe* today's behavior on first load rather than silently changing it:

| View | Default `sortBy` | Default `sortDir` | Reproduces |
|---|---|---|---|
| Folder | `manual` | n/a | `position ASC` (today's only folder ordering) |
| All / Unsorted / Trash | `created_at` | `desc` | `created_at DESC` (today's only non-folder ordering) |

Switching views resets to the new view's default — identical to how `searchTerm` clears on navigation (`AppLayout.tsx:93`), so sort doesn't "leak" a folder's explicit choice into All, or vice versa.

**Alternative considered**: let sort choice persist across view switches (e.g., user picks "Name" in All, then opens a folder and still sees "Name"). Rejected per your direction — "since filter by name resets upon switching folder, sort should be the same." Consistency with the existing reset behavior also avoids a subtle bug class: a folder-scoped `Manual` selection wouldn't have an equivalent in All, so *some* reset rule is unavoidable; resetting on every switch is the simplest rule that's already proven out for search.

### 4. Direction toggle is conditionally rendered, not just disabled, when `sortBy === 'manual'`

The `↑`/`↓` row (with field-specific labels — "Oldest/Newest first" vs "A→Z"/"Z→A", per the handoff's `DIR_LABELS`) is omitted entirely from the panel when `Manual` is selected, rather than shown-but-disabled. Direction has no meaning for manual order — showing a disabled, label-less control would raise more questions ("disabled toggle for what?") than it answers. This also sidesteps needing a `DIR_LABELS['manual']` entry.

### 5. Reorder-specific gates gain `&& sortBy === 'manual'`; pickup itself stays untouched

A single `useSortable` on `ImageCard` (`ImageGrid.tsx:51-55`) backs **two** distinct drag outcomes, routed by `AppLayout`'s top-level `DndContext.onDragEnd` (`AppLayout.tsx:125-164`) based on what the card is dropped on:

```
pick up an image card (useSortable — ONE drag source, gated only by `disabled: isTrash`)
        │
        ├─ dropped on another image  → sortEndTrigger → reorder via position  (folder + manual only)
        │
        └─ dropped on a folder       → handleImageDrop → move image into folder  (always, except trash)
```

Folder-restructure (drag image → folder) already works in All/Unsorted today *without* any `SortableContext` wrap — those views render the plain `grid` (line 331-332), and pickup still works because `useSortable`'s underlying `useDraggable` doesn't require a `SortableContext` to be draggable; the routing happens at the `DndContext` level. That existing behavior is exactly the target shape for "folder view, non-manual sort": **identical to All/Unsorted**, not a new restricted mode.

So the corrected extension touches only the *reorder-specific* call sites — `disabled` is left as `isTrash` alone:

| Gate | Current | Corrected | Effect |
|---|---|---|---|
| `useSortable({ disabled })` (line 53) | `isTrash` | **unchanged** | Pickup stays available outside trash → folder-restructure drag keeps working in every non-trash context, matching All/Unsorted |
| `SortableContext` wrap (lines 327-330) | `isFolderView` | `isFolderView && sortBy === 'manual'` | Non-manual folder view renders the plain `grid`, becoming behaviorally identical to All/Unsorted for DnD purposes |
| `onDragOver` reorder-preview (line 152) | `if (!isFolderView \|\| ...) return` | add `\|\| sortBy !== 'manual'` | The reorder drop-indicator ring only appears in manual mode |
| `sortEndTrigger` effect (line 237) | `if (... view.type !== 'folder') return` | add `\|\| sortBy !== 'manual'` | Position is only ever persisted in manual mode |

Resulting matrix:

| | Folder + Manual | Folder + explicit sort | All / Unsorted | Trash |
|---|---|---|---|---|
| `useSortable` enabled | ✓ | ✓ | ✓ | ✗ (`isTrash`) |
| `SortableContext` wrapped | ✓ | ✗ | ✗ | — |
| reorder → persists `position` | ✓ | ✗ | ✗ | ✗ |
| drag → move to folder | ✓ | ✓ | ✓ | ✗ |

**Why not gate `useSortable`'s `disabled` on `sortBy !== 'manual'`** (my first instinct, corrected after walking through the drop-routing): that flag suppresses `listeners`/`attributes` entirely, killing pickup — which would also kill folder-restructure drag in a non-manual-sorted folder. That makes the folder view *more* restrictive than All/Unsorted under explicit sort, which directly contradicts "not manual sort in folder → same as all/unsorted behaviour." Gating only the `SortableContext` wrap and the two reorder-persistence call sites achieves exactly that equivalence, while the always-on `useDraggable` (via `useSortable({ disabled: isTrash })`) is what already makes folder-restructure drag possible in All/Unsorted — reusing it, unmodified, is what makes folder-view-with-explicit-sort behave the same way "for free."

**Alternative considered**: leave reorder enabled under explicit sort, and let the next fetch silently override any dragged position. Rejected — that's a worse UX (the dragged card would visually "snap back" on refetch) and contradicts the stated intent (disable, Eagle-style) to avoid a confusing, self-undoing interaction.

### 6. Sort control reuses `DropdownMenu` + `DropdownMenuRadioGroup`/`DropdownMenuRadioItem`, not a bespoke panel

The design handoff's `Panel`/`RadioDot`/`SortPanelBody` are throwaway prototype components with inline styles. The actual control should be built from primitives this codebase already has and uses elsewhere — `DropdownMenu`/`DropdownMenuTrigger`/`DropdownMenuContent` (already used for the upload split-button, `AppLayout.tsx:253-267`) wrapping a `DropdownMenuRadioGroup` of `DropdownMenuRadioItem`s for the field choice, plus a plain item or inline control for the direction toggle. The trigger is an icon-only button using `buttonVariants`, matching the handoff's visual intent (a small square icon button left of/near the search input) without introducing a new panel/menu pattern alongside the existing `DropdownMenu` one.

**Alternative considered**: port the handoff's bespoke `Panel`/`RadioDot` components as-is. Rejected — that would introduce a second floating-panel implementation alongside `DropdownMenu`, duplicating open/close/positioning/focus-trap logic the existing primitive already handles, and would need `asChild`-free Base UI patterns reverified from scratch (per [[feedback_no_aschild]] — Base UI, not Radix).

The trigger also visually distinguishes an active (non-default) sort, mirroring the handoff's `IconBtn`'s `active` prop (`color: active ? FG : MFG`, `Filter & Sort Options.html:72-96`) — but the comparison target differs: the handoff checks against one global default (`sortBy !== 'date_added' || sortDir !== 'desc'`), whereas ours must check against the *current view's* default (Decision #3's table), since `manual`-in-a-folder and `created_at`/`desc`-in-All are both "default, not active" states:

```
sortActive = sortBy !== viewDefaultSortBy
          || (sortBy !== 'manual' && sortDir !== viewDefaultSortDir)
```

`sortActive` is purely derived from state already being tracked (`sortBy`/`sortDir` vs. the per-view default) — no new state. Visually, rather than porting the handoff's badge-dot (which has no precedent anywhere in this codebase — `RightPanel`/`ImageViewer`'s absolute-positioned elements serve unrelated purposes), the trigger conditionally switches `buttonVariants` styling when `sortActive` is true, reusing the same variant mechanism the trigger is already built from rather than introducing a new indicator pattern.

## Risks / Trade-offs

- **[Risk]** Adding `sort`/`direction` to `ImageGrid`'s query key (`queryKeyFor`, lines 103-110) means every sort change triggers a fresh paginated fetch — for large folders/views this could feel like a "reload" rather than an instant re-sort. → **Mitigation**: this matches how `debouncedSearchTerm` already works (also in the query key, also triggers refetch with `keepPreviousData`); `useInfiniteQuery`'s `placeholderData: keepPreviousData` (line 180) keeps the previous page visible during the transition, so there's no jarring blank state.
- **[Trade-off]** `Manual` only existing in folder views means the radio list has different membership per view type — slightly more conditional rendering than a single static list. → Accepted: this mirrors a real asymmetry the backend already encodes (`position` is folder-scoped; non-folder views have no concept of manual order), so hiding `Manual` elsewhere is honest about what's actually orderable, not an arbitrary UI restriction.
- **[Trade-off]** Disabling drag under explicit sort removes a capability (reordering) the user had a moment ago, which could surprise someone mid-drag if they change sort via another input. → Accepted: sort and drag are both single-focus interactions (you can't open the sort panel while mid-drag), so this scenario can't actually occur in practice; the disable takes effect cleanly between interactions.

## Migration Plan

No backend or database changes — purely additive frontend work consuming an existing, already-shipped API contract. Deployable as a single frontend change: add the toolbar control, thread `sort`/`direction` through `images.ts`/`ImageGrid`, extend the three (now four, including `SortableContext`) drag-gating conditions, and verify against the running backend (which already supports the params). Rollback is a plain revert — no persisted state changes shape.

## Open Questions

None outstanding — the backend contract is settled and archived, the manual-vs-explicit-sort scoping was resolved in this exploration (folder-only `Manual`, reset-on-switch, disable-drag-when-not-manual), and the remaining sort fields (`file_size`/`dimensions`) are explicitly deferred pending backend support.
