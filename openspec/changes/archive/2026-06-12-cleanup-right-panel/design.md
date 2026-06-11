## Context

`RightPanel.tsx` (288 lines) hosts `ImagePanelBody`, which mixes several
field-editing concerns (title/notes/source/folders/tags), a details summary,
and a download action, plus two file-local formatting helpers. A sibling
mode, `FolderPanelContent`, hand-rolls its own autosave logic instead of
reusing `useFieldAutosave`. `FolderInput` and `TagInput` are two
near-identical chip-input implementations. `useVisionSuggestion` lives under
`right-panel/hooks/` but is only used by `AppLayout`.

All five changes are zero-functional-change extractions/relocations — no new
state, no new network calls, no prop changes on `RightPanel` itself.

## Goals / Non-Goals

**Goals:**
- Reduce `RightPanel.tsx` to its orchestration role (mode dispatch + the
  field sections that are genuinely specific to image-detail editing).
- Remove the duplicated autosave logic in `FolderPanelContent`.
- Replace `FolderInput`/`TagInput`'s duplicated chip-input implementation
  with one shared component, without changing either component's external
  props or behavior.
- Move `useVisionSuggestion` to the feature that actually owns it.

**Non-Goals:**
- No change to `RightPanel`'s props, `AppLayout`'s usage, or any rendered
  markup/styling.
- No change to tag free-text creation, folder/tag save behavior, or the
  vision-suggestion flow itself.
- Not pursuing the deferred `GalleryQuery`/`useGalleryImages` param-object
  cleanup (separate thread, out of scope here).

## Decisions

### D1: `TokenInput<T>` shape

`FolderInput` and `TagInput` differ in three places only:

| | `FolderInput` | `TagInput` |
|---|---|---|
| Item type | `Folder { id, name }` | `Tag { id, name }` |
| Free-text creation | none | `commitRaw` (comma key, blur-commit, lowercases + strips commas, skips duplicates by name) |
| Placeholder | `"Add to folder…"` | `"Add tags…"` |
| Remove matching | by `id` | by `id`, falling back to `name` if `id` is empty |

Everything else (state: `val`/`selectedIndex`/`showDropdown`, the
`blurTimerRef` 150ms close delay, dropdown filtering, arrow/enter/escape/
backspace keyboard handling, chip rendering) is identical.

`TokenInput<T extends { id: string; name: string }>` will own all of that
shared behavior:

```ts
interface TokenInputProps<T extends { id: string; name: string }> {
  items: T[]
  onChange: (items: T[]) => void
  disabled?: boolean
  suggestions?: T[]
  placeholder?: string
  // When provided, enables free-text entry: comma key and blur commit a raw
  // string via this function. Returning null/undefined means "don't add"
  // (e.g. empty after trim, or duplicate name) — mirrors TagInput's
  // commitRaw guard clauses.
  createFromText?: (raw: string) => T | null
}
```

- Removal is done by item reference (`items.filter((i) => i !== target)`)
  rather than by id/name matching — this drops the `id || name` fallback
  `TagInput.remove` needs today, since `TokenInput` always has the actual
  item object (from `items` or from `createFromText`'s return) to filter on.
  Chip `key` uses `item.id || item.name` (unchanged from today, needed for
  freshly-created tags that may render before `id` is resolved).
- `FolderInput` and `TagInput` remain as thin wrapper files with their
  current prop names (`folders`/`tags`, `FolderInputProps`/`TagInputProps`)
  — they configure `TokenInput` with their type, placeholder, and (for
  `TagInput`) `createFromText`. `RightPanel.tsx` and its imports/usage of
  `FolderInput`/`TagInput` do not change.

**Alternative considered**: have `RightPanel` import `TokenInput` directly
twice, dropping `FolderInput`/`TagInput` entirely. Rejected — it would touch
`RightPanel.tsx` (already shrinking via other parts of this change) and
their existing test files, for no behavioral benefit over a thin wrapper.

### D2: `FolderPanelContent` → `useFieldAutosave`

Maps directly onto the existing title/description pattern already used in
`ImagePanelBody`:

- `name` field: `useFieldAutosave(folder.name, (value) => saveMutation.mutate({ name: value }), { isEmpty: (v) => v.trim() === '' })`
  — same revert-if-empty rule as the image title field.
- `description` field: `useFieldAutosave(folder.description ?? '', (value) => saveMutation.mutate({ description: value || null }))`
  — no `isEmpty`, matching the image description field.

One nuance: today's `FolderPanelContent` resyncs local state on
`[folder.id, folder.name, folder.description]` (i.e. whenever the *folder*
changes, even if name/description happen to be unchanged strings).
`useFieldAutosave` resyncs only when its `value` argument changes. In the
edge case of switching between two folders that happen to share the exact
same name/description string, `useFieldAutosave`'s effect wouldn't re-fire
(value is unchanged) — but since `local` already equals that shared string,
there's nothing to visibly reset. To be safe and keep the remount semantics
identical, `RightPanel` will render `<FolderPanelContent key={props.folder.id} ... />`,
so switching folders always remounts the component and re-initializes both
`useFieldAutosave` instances from scratch — matching today's per-folder
reset exactly.

### D3: Sequencing

Unlike `split-imagegrid-concerns`, none of these five changes depend on each
other:

- `DetailsGrid` extraction, `DownloadButton` extraction, `FolderPanelContent`
  autosave swap, `TokenInput` unification, and `useVisionSuggestion`
  relocation each touch disjoint files (aside from `RightPanel.tsx`, which
  only loses code in the first two).
- They can be implemented and tested in any order, and `tasks.md` will list
  them as independent groups rather than a numbered pipeline.

### D4: `useVisionSuggestion` relocation

Pure file move: `features/right-panel/hooks/useVisionSuggestion.ts` (and its
test) → `app-shell/useVisionSuggestion.ts`. No internal changes; only
`AppLayout.tsx`'s import path is updated.

## Risks / Trade-offs

- **[Risk]** `TokenInput`'s `createFromText` generalization could subtly
  change `TagInput`'s exact normalization/dedup behavior if not ported
  carefully. → **Mitigation**: `TagInput`'s wrapper passes a
  `createFromText` that reproduces `commitRaw`'s existing logic verbatim
  (lowercase, strip commas, trim, dedup-by-name against current `tags`);
  existing `TagInput.test.tsx` scenarios continue to assert this behavior
  through the wrapper.
- **[Risk]** Reference-based removal in `TokenInput` assumes the `items`
  array passed in always contains the actual objects to be removed (true
  for both `selectedFolders` and `tags` today, since both are plain arrays
  rebuilt from query data / `onChange`). → **Mitigation**: covered by
  existing FolderInput/TagInput remove tests (chip ✕ button and Backspace).
- **[Trade-off]** Adding `key={props.folder.id}` to `FolderPanelContent` in
  `RightPanel.tsx` is a one-line change to a file this proposal otherwise
  only shrinks. It's necessary to preserve exact reset semantics under D2
  and is small enough not to warrant its own change.
