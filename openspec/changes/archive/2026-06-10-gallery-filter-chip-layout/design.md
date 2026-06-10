## Context

The "Add filter" panel lives inside `AppLayout.tsx`'s existing `DropdownMenu`/`DropdownMenuContent` (Base UI Menu). Today it renders three `DropdownMenuGroup`s (Tags, File type, Folder) back-to-back as `DropdownMenuCheckboxItem` rows, relying on the panel's own `max-h-(--available-height) overflow-y-auto` for scrolling. The "Filter & Sort Options" design handoff (Option C) groups these into clearly divided sections, with File type as a row of toggle chips and Tags/Folder as searchable lists.

A spike into Base UI's `Menu` internals (`floating-ui-react`'s `useTypeahead`) found that while the menu is open, any single-character keydown bubbling to the popup is `preventDefault()`'d for typeahead-to-focus — with no built-in exemption for `<input>` elements. A plain search `<input>` inside `DropdownMenuContent` would therefore have its keystrokes silently swallowed unless this is worked around.

## Goals / Non-Goals

**Goals:**
- Reorganize the filter panel into three sections (File type, Tags, Folder), each with an `InlineDivider`-style header.
- Render File type as a multi-select toggle/chip row using new shared `ui/toggle.tsx` + `ui/toggle-group.tsx` primitives (shadcn `base-nova`).
- Add a search input to the Tags and Folder sections that filters their list by case-insensitive substring match on name; empty results hide the section's list entirely (including checked items).
- Give Tags and Folder lists their own fixed `max-h` + `overflow-y-auto`, independent of the panel's own scroll.
- Keep the panel inside the existing `DropdownMenu`/`DropdownMenuContent` — no new floating/popover primitive.

**Non-Goals:**
- No change to the toolbar layout, "Filters" button/badge, active-filter chip row, or `clearAllFilters`.
- No change to which sections appear per view (`filterSectionsForViewType`) or to query-param wiring (`tag_ids`, `mime_types`, `folder_ids`).
- No `ui/badge.tsx` — active-filter pills keep their current hand-rolled markup.
- No autofocus on the new search inputs.
- No "+N more" / overflow affordance for File type — it remains the fixed 5-value set, just rendered as toggles instead of checkboxes.

## Decisions

### File type as `ToggleGroup` (`type="multiple"`)
Pull `ui/toggle.tsx` and `ui/toggle-group.tsx` from the shadcn `base-nova` registry (wrapping `@base-ui/react/toggle` and `@base-ui/react/toggle-group`, already present as transitive deps). Render the 5 `MIME_TYPE_FILTER_OPTIONS` as a `ToggleGroup type="multiple"` whose `value`/`onValueChange` map directly to `filterMimeTypes: string[]` — no intermediate state needed. Toggle styling is adjusted (pill/`rounded-full`) to match the design handoff's `Chip` look.

**Alternative considered**: hand-rolled `Chip` button per option (as in the design handoff's raw HTML). Rejected — shadcn already ships an equivalent primitive for this style, and reusing it avoids a one-off component for a 5-item list.

### Search input inside `DropdownMenuContent`, with `stopPropagation` on `keydown`
Add a plain `<input>` (styled like the existing toolbar search) above each of the Tags and Folder lists. Its `onKeyDown` calls `event.stopPropagation()` so the keystroke never reaches `MenuPopup`'s `onKeyDown`, where `useTypeahead` would otherwise call `preventDefault()` on it. Arrow-key/Escape behavior for the rest of the menu (handled via `useListNavigation`/`useDismiss`, the latter attached at `document` level) is unaffected.

**Alternative considered**: migrate the filter panel to `@base-ui/react/popover` (not currently wrapped in `src/components/ui`), which has no typeahead and would sidestep the issue entirely. Rejected for this change — it's a new UI primitive/pattern, larger blast radius, and the `stopPropagation` fix is small, scoped, and keeps the panel on the primitive the existing `fe-gallery-filters` spec already mandates.

### Filtering behavior: local state, reset on view change
Add `filterTagSearch: string` and `filterFolderSearch: string` to `AppLayout`'s existing filter state, reset alongside `filterTagIds`/`filterMimeTypes`/`filterFolderIds` on view change (same effect that already clears those). The displayed list for each section is `items.filter(i => i.name.toLowerCase().includes(search.toLowerCase()))`; when `search` is non-empty and the filtered list is empty, render nothing below the search input (not even currently-checked items) — per the proposal's explicit empty-state behavior.

### Per-section scroll caps
Wrap each of the Tags and Folder checkbox lists in their own `<div className="max-h-40 overflow-y-auto">` (exact height TBD during implementation/visual review). This nests inside the panel's existing `max-h-(--available-height) overflow-y-auto` without conflict — nested `overflow-y-auto` regions scroll independently once their own content exceeds their cap.

## Risks / Trade-offs

- **[Risk]** `stopPropagation` on the search input's `keydown` could mask other menu-level keyboard behavior while the input is focused (e.g., arrow-key navigation between checkbox items won't work from inside the input). → **Mitigation**: this is the desired behavior — while typing, arrow keys should move the text cursor, not menu focus; users can `Tab`/click out of the input to resume list navigation.
- **[Risk]** If a section's `max-h` is set too generously, the panel could still need its own outer scroll, producing nested scrollbars. → **Mitigation**: pick a conservative cap (e.g. ~6-7 rows) and verify visually during implementation against the tallest realistic section (Tags).
- **[Risk]** shadcn's default `toggle`/`toggle-group` styling (likely `rounded-md`) doesn't match the design handoff's pill (`rounded-full`) chips. → **Mitigation**: override via `className`/variant when composing the File type row; no change to the shared component's default variants needed elsewhere.

## Migration Plan

1. Pull `ui/toggle.tsx` and `ui/toggle-group.tsx` via the shadcn CLI.
2. Rework the "Add filter" panel JSX in `AppLayout.tsx`: section dividers, File type as `ToggleGroup`, Tags/Folder with search input + capped scroll list.
3. Add `filterTagSearch`/`filterFolderSearch` state and reset logic.
4. Update `AppLayout.test.tsx` for the new section structure and toggle-based file type interaction.

No backend changes, no data migration, no feature flag — this is a pure frontend presentational/interaction change behind the existing "Filters" button. Rollback is a straightforward revert of the FE commit(s).

## Open Questions

- Exact `max-h` value (in rows/px) for the Tags and Folder scroll regions — to be tuned visually during implementation.
- Visual styling details for the pill-shaped `ToggleGroup` items (padding, active/inactive colors) — should match existing `--secondary`/`--foreground` tokens per the design handoff's token mapping, finalized during implementation.
