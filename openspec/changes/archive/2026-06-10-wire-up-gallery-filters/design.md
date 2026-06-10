## Context

The backend contract already exists (`advanced-gallery-filters`, archived): `GET /images` accepts optional `folder_ids`, `tag_ids`, `mime_types` as comma-separated, match-any (`IN`/`EXISTS`) filters that compose with each other and with `name`/`unfiled`/`sort`/pagination via `AND`. `unfiled=true` (used by the Unsorted view, `getImages`) combined with any `folder_ids` is a contradiction that yields an empty result — no validation prevents it, it just "answers literally." This phase only has to *consume* that contract — no new backend work.

Today's toolbar (`AppLayout.tsx:293-333`) is:

```tsx
<div className="flex items-center justify-between gap-2 mb-4">
  <div className="flex items-center gap-2 w-full max-w-xs">
    {/* search input, flex-1 */}
    {/* sort DropdownMenu (icon button) */}
  </div>
  <div className="flex">{/* upload split-button */}</div>
</div>
```

`ImageGrid.tsx` builds its query key/fetcher per view (`queryKeyFor`/`fetcherFor`, lines 111-131) and already does client-side filtering for folder-view search (lines 196-199):

```tsx
const allImages = isFolderView && trimmedSearch
  ? fetchedImages.filter((img) => img.title.toLowerCase().includes(trimmedSearch))
  : fetchedImages
```

`Image` (`lib/images.ts`) already carries `tags: { id; name }[]` and `mime_type` on every item, so folder-view client-side filtering needs no extra fetch.

The design handoff ("Filter & Sort Options.html", Option C) shows a "Filters" text button (badge = active count) opening a panel with `Chip`/`CheckRow` sections for Tags / File type / Folder, plus an active-filter chip row below the toolbar. Like the sort handoff, this is a bespoke prototype (`Panel`/`Chip`/`CheckRow`/`InlineDivider`) not bound to this app's component library.

The codebase's `dropdown-menu.tsx` already exports `DropdownMenuCheckboxItem` (Base UI `Menu.CheckboxItem`) — unused so far, but exactly the multi-select toggle primitive Option C's `Chip`/`CheckRow` sections need, alongside `DropdownMenuLabel`/`DropdownMenuSeparator` for section headers (`fe-gallery-sort` already established the precedent of building toolbar panels from `DropdownMenu*` primitives rather than the handoff's bespoke `Panel`).

`TagInput.tsx:113` has the only existing "pill" styling in the codebase (`inline-flex items-center gap-1 bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs`), used for tags attached to an image.

Accepted upload mime types (`UploadModal.tsx:22`, `BatchUploadModal.tsx:20`): `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/avif`, `image/heic` — but `image/heic` is converted to `image/jpeg` before upload (`UploadModal.tsx:119-121`), so it never appears as a stored `mime_type`. The handoff's `ALL_TYPES` (`JPEG/PNG/GIF/WEBP/SVG`) includes SVG, which isn't an accepted upload type and can't appear in the data.

## Goals / Non-Goals

**Goals:**
- Add a "Filters" button + panel (Option C) to the gallery toolbar, multi-select across Tags / File type / Folder, with an active-filter chip row.
- Scope filter sections per view: All gets all three; Unsorted and Folder view get Tags + File type only (Folder also drops it for being meaningless, Unsorted because `folder_ids` + `unfiled=true` is a contradiction); Trash hides the control entirely.
- Folder view filtering is client-side over the already-fetched full image list, extending the existing search-filter pattern — no new endpoint call.
- Reset filter selections to empty on view switch, mirroring `searchTerm`/sort reset.
- Build the panel from existing `DropdownMenu*` primitives (`DropdownMenuCheckboxItem`, `DropdownMenuLabel`, `DropdownMenuSeparator`), and active-filter chips from the existing `TagInput` pill style — no new floating-panel or pill pattern.
- File type options are a fixed, accurate list derived from mime types that can actually exist on stored images.

**Non-Goals:**
- A visible "active sort" pill/chip — out of scope per the proposal; the sort button's existing active-variant styling is untouched.
- Match-ALL ("has tag A AND tag B") semantics — the backend only supports match-any, and the panel should not imply otherwise.
- Persisting filter selections across views or sessions.
- Options A/B layouts from the design handoff.
- Adding `image/svg+xml` (or any type not in `UploadModal`'s accepted set) as a filter option.

## Decisions

### 1. Filter state lives in `AppLayout`, passed to `ImageGrid` as props — same lifecycle as `searchTerm`/`sortBy`

Three pieces of state: `filterTagIds: string[]`, `filterMimeTypes: string[]`, `filterFolderIds: string[]`. All declared in `AppLayout`, reset to `[]` in the existing `viewKey`-keyed `useEffect` that already resets `searchTerm`/`sortBy`/`sortDir` (`AppLayout.tsx:126-134`). `ImageGrid` receives them as props purely to build its query key/fetcher (All/Unsorted) or apply a client-side `.filter()` (Folder) — it doesn't own the selection UI, identical to how `sortBy`/`sortDir` are split (`wire-up-image-sort` Decision #2).

**Alternative considered**: scope `filterFolderIds` only to the All view (since it's the only view that uses it) by not declaring it for other views. Rejected — a single flat state shape that's simply unused/empty for other views is simpler than view-conditional state shapes, and the reset effect already clears it uniformly.

### 2. Toolbar layout: wrap the existing `max-w-xs` group, add Filters as a sibling; chip row as a new full-width row below

```tsx
<div className="flex flex-col gap-2 mb-4">
  <div className="flex items-center justify-between gap-2">
    <div className="flex items-center gap-2">
      <div className="flex items-center gap-2 w-full max-w-xs">
        {/* search input (flex-1), sort DropdownMenu — unchanged */}
      </div>
      {view.type !== 'trash' && <FiltersButton />}  {/* outside max-w-xs */}
    </div>
    <div className="flex">{/* upload split-button — unchanged */}</div>
  </div>
  {activeChips.length > 0 && <ActiveFilterChipRow />}
</div>
```

This is a minimal structural change: the search+sort `max-w-xs` div is wrapped in a new `flex items-center gap-2` div alongside the Filters button (so search keeps its width cap, Filters doesn't), and the whole toolbar row is wrapped in a `flex-col` so the chip row can sit below it without disturbing `justify-between`. The chip row renders conditionally — no reserved space when empty, per the proposal.

**Alternative considered**: put Filters inside the existing `max-w-xs` div (closest to the handoff's literal layout). Rejected per your direction — "outside max-w-xs is fine" — and a text button with a count badge needs more breathing room than the 320px cap allows alongside a usable search input.

### 3. Filter panel built from `DropdownMenu` + `DropdownMenuCheckboxItem`, sectioned with `DropdownMenuLabel`/`DropdownMenuSeparator`

```tsx
<DropdownMenu>
  <DropdownMenuTrigger className={cn(buttonVariants({ variant: filterCount > 0 ? 'default' : 'outline' }), '...')}>
    <Filter className="w-3.5 h-3.5" /> Filters {filterCount > 0 && <Badge>{filterCount}</Badge>}
  </DropdownMenuTrigger>
  <DropdownMenuContent align="start" className="w-56">
    <DropdownMenuLabel>Tags</DropdownMenuLabel>
    {tags.map(t => <DropdownMenuCheckboxItem checked={...} onCheckedChange={...}>{t.name}</DropdownMenuCheckboxItem>)}
    <DropdownMenuSeparator />
    <DropdownMenuLabel>File type</DropdownMenuLabel>
    {MIME_TYPE_OPTIONS.map(...)}
    {view.type === 'all' && <>
      <DropdownMenuSeparator />
      <DropdownMenuLabel>Folder</DropdownMenuLabel>
      {folders.map(...)}
    </>}
  </DropdownMenuContent>
</DropdownMenu>
```

This mirrors `fe-gallery-sort`'s Decision #6 (reuse `DropdownMenu*`, not the handoff's bespoke `Panel`/`Chip`/`CheckRow`). The badge count uses the same `buttonVariants` active/inactive switch the sort trigger already uses (`sortActive ? 'default' : 'outline'`), extended with a small count badge — no new indicator pattern beyond adding a number, matching the handoff's `IconBtn`/Option C `Filters` button intent.

**Open question carried to tasks**: Base UI's `Menu.CheckboxItem` close-on-select behavior needs verification during implementation — the sort panel needed `closeOnClick={false}` on its direction-toggle `DropdownMenuItem` (`AppLayout.tsx:325`) to stay open, but `CheckboxItem` may default differently (checkbox items conventionally don't auto-close, to support multi-select). If it *does* close on toggle, each `DropdownMenuCheckboxItem` here will need the same `closeOnClick={false}` (or equivalent) so users can toggle multiple filters without reopening the panel each time.

### 4. File type options: fixed list derived from actual stored `mime_type` values, not the handoff's `ALL_TYPES`

A new constant (in `lib/images.ts`, alongside the other `Image`-related types/constants) maps the mime types that can actually appear on a stored image to display labels:

```ts
export const MIME_TYPE_FILTER_OPTIONS: { value: string; label: string }[] = [
  { value: 'image/jpeg', label: 'JPEG' },
  { value: 'image/png',  label: 'PNG'  },
  { value: 'image/gif',  label: 'GIF'  },
  { value: 'image/webp', label: 'WEBP' },
  { value: 'image/avif', label: 'AVIF' },
]
```

This excludes `image/heic` (converted to `image/jpeg` pre-upload, per `UploadModal.tsx:119-121` — never a stored value) and the handoff's `SVG` (not an accepted upload type at all). If `UploadModal`'s accepted-type set changes in the future, this list and that set could drift — acceptable for now since both are small, manually-curated lists with no shared source of truth today.

**Alternative considered**: derive the file-type filter options dynamically from the distinct `mime_type` values present in the user's images (a new endpoint/aggregation). Rejected as out of scope — no such endpoint exists, and a static list matching the upload-accepted set is accurate enough (a user can only ever have images of types they were able to upload).

### 5. Query params: `tagIds`/`mimeTypes`/`folderIds` → comma-separated `tag_ids`/`mime_types`/`folder_ids`, mirroring backend's CSV decision

`getImages`/`getAllImages` (`src/lib/images.ts`) gain optional `tagIds?: string[]`, `mimeTypes?: string[]`, `folderIds?: string[]` (the last only meaningful for `getAllImages`, but added to both for signature symmetry — `getImages` simply won't be called with it from the Unsorted view). Each, when non-empty, is joined with `,` and set as `tag_ids`/`mime_types`/`folder_ids` — matching the backend's CSV encoding decision (`advanced-gallery-filters` design Decision #3) and the existing `URLSearchParams` usage pattern in these functions.

`ImageGrid`'s `queryKeyFor`/`fetcherFor` (lines 111-131) thread the three arrays through for `'all'`/`'unsorted'`, the same way `sort`/`direction` were threaded in `wire-up-image-sort`.

### 6. Folder view: extend the existing client-side filter to also match tags and mime type

`ImageGrid.tsx`'s existing folder-view filter (lines 196-199) becomes:

```tsx
const allImages = isFolderView
  ? fetchedImages.filter((img) =>
      (!trimmedSearch || img.title.toLowerCase().includes(trimmedSearch)) &&
      (filterTagIds.length === 0 || img.tags.some(t => filterTagIds.includes(t.id))) &&
      (filterMimeTypes.length === 0 || filterMimeTypes.includes(img.mime_type))
    )
  : fetchedImages
```

Each filter dimension is independently optional (empty array = no constraint) and combines via `&&` — mirroring the backend's "AND across filters, OR within each" semantics (`advanced-gallery-filters` design Decision #5), just evaluated client-side. `filterFolderIds` is never passed for folder view (Decision #1 — always `[]` there), so it never participates here.

### 7. Active-filter chips reuse `TagInput`'s pill styling; one chip per selected value across all three dimensions

```tsx
const activeChips = [
  ...selectedTags.map(t => ({ key: `tag:${t.id}`, label: t.name, onRemove: () => toggleTag(t.id) })),
  ...selectedMimeTypes.map(m => ({ key: `mime:${m}`, label: MIME_LABEL[m], onRemove: () => toggleMime(m) })),
  ...selectedFolders.map(f => ({ key: `folder:${f.id}`, label: f.name, onRemove: () => toggleFolder(f.id) })),
]
```

Each chip: `inline-flex items-center gap-1 bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs` (verbatim from `TagInput.tsx:113`) plus a small `×`/`X` remove button (`TagInput.tsx:123-126`'s `<X className="w-2.5 h-2.5" />` pattern). A trailing "Clear all" text button resets all three arrays to `[]`. Folder/tag chips need name lookups against the already-available `getFolders()`/`getTags()` results; mime chips use `MIME_TYPE_FILTER_OPTIONS`'s label map.

**Alternative considered**: port the handoff's `Chip` component (rounded-pill, `border-radius:20`, active/inactive fill) for active-filter chips. Rejected — would introduce a second pill style alongside `TagInput`'s existing one for what is visually the same concept (a tag/label with a remove affordance).

## Risks / Trade-offs

- **[Risk]** `DropdownMenuCheckboxItem`'s close-on-select behavior is unverified (see Decision #3) — if it auto-closes, every checkbox needs an explicit override, which is mechanical but touches every option across three sections. → **Mitigation**: verify against one item early in implementation before wiring all three sections.
- **[Trade-off]** The chip row appearing/disappearing shifts the grid's vertical position by one row's height. → Accepted per prior discussion — this is the explicit trade for Option C's visibility, and is a common, recognizable pattern (Linear/Notion-style filter bars).
- **[Trade-off]** `MIME_TYPE_FILTER_OPTIONS` and `UploadModal`'s `ACCEPTED_TYPES` are two manually-maintained lists that could drift if upload support changes. → Accepted — both are small and rarely change; introducing a shared constant module for two three-to-six-element lists would be premature abstraction for this change.
- **[Trade-off]** `folderIds` is plumbed through `getImages` (Unsorted) for signature symmetry with `getAllImages` even though it's never populated there. → Accepted — keeps the two functions' signatures parallel (they already share `cursor`/`name`/`sort`/`direction`), avoiding a special-cased signature for one view.

## Migration Plan

No backend or database changes — purely additive frontend work consuming an already-shipped API contract (`advanced-gallery-filters`). Deployable as a single frontend change: add the toolbar control + chip row, thread `tag_ids`/`mime_types`/`folder_ids` through `images.ts`/`ImageGrid` for All/Unsorted, extend folder view's client-side filter, and verify against the running backend (which already supports the params). Rollback is a plain revert — no persisted state changes shape.

## Open Questions

- `DropdownMenuCheckboxItem` close-on-select behavior (Decision #3) — to be confirmed during implementation; resolution is mechanical (`closeOnClick={false}` or equivalent) and doesn't change the design's shape either way.
