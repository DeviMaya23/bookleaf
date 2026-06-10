## Why

The gallery's "Filters" panel currently renders Tags, File type, and Folder as one long vertical stack of checkboxes in a single dropdown. As a user's tag and folder lists grow, this becomes a long scrollable list that's hard to scan. The "Filter & Sort Options" design handoff (Option C: inline filter chip bar) reorganizes this into clearly separated sections with better visibility for each filter type.

## What Changes

- Reorganize the "Add filter" dropdown panel into three sections, each with an inline divider header: **File type**, **Tags**, **Folder**.
- **File type** becomes a multi-select toggle/chip row (pill-shaped toggles) instead of a checkbox list, built from new shared `ui/toggle.tsx` and `ui/toggle-group.tsx` components (pulled from the shadcn `base-nova` registry).
- **Tags** and **Folder** sections each gain a search input above their checkbox list that filters the list by case-insensitive name match. When the search yields no matches, the section shows no items at all (including already-checked ones).
- **Tags** and **Folder** checkbox lists each become independently height-capped (`max-h-*`) and scrollable (`overflow-y-auto`), instead of relying on the dropdown panel's own scroll.
- The new search inputs call `event.stopPropagation()` on `keydown` to prevent Base UI Menu's typeahead behavior from intercepting and blocking keystrokes.
- No change to: the toolbar layout, the "Filters" button and its count badge, the active-filter chip row below the toolbar, which sections appear per view type, or how selected filters are sent to/applied against the image list.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `fe-gallery-filters`: the "Filters control in gallery toolbar" requirement's panel description changes (sectioned layout with dividers, toggle-group for file type); the "Tag and folder options are sourced from existing data" requirement gains search/filter and scroll-cap behavior, including the empty-search-result behavior.

## Impact

- `frontend/src/components/AppLayout.tsx` — filter panel JSX and related state (search terms for tags/folders)
- `frontend/src/components/ui/toggle.tsx`, `frontend/src/components/ui/toggle-group.tsx` — new shared UI primitives (pulled via shadcn CLI, not hand-written)
- `frontend/src/components/AppLayout.test.tsx` — existing filter-panel tests need updating for the new section layout and toggle-based file type interaction
- No backend or API changes; `tag_ids`/`mime_types`/`folder_ids` query parameter contract is unchanged
