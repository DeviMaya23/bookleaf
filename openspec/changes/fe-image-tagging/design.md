## Context

The backend ships `GET /tags`, `POST /tags`, and `PATCH /images/:id` (with `tags: uuid[]` replace). Image list and detail responses already include `tags: [{id, name}]`. The frontend `Image` type and `RightPanel` are unaware of tags. The right panel currently auto-saves title, notes, and source URL on blur — tags need to follow the same pattern.

## Goals / Non-Goals

**Goals:**
- Render a `TagInput` section in the right panel between Source and Details
- Load and cache all user tags via `GET /tags` so name→ID resolution is local
- On each tag add/remove, fire `PATCH /images/:id` with the full updated tag UUID set
- When adding a new tag name (not in cache), create it via `POST /tags` first, then patch

**Non-Goals:**
- Filter-by-tag UI anywhere in the app
- Tag renaming or deletion from the frontend
- Autocomplete dropdown / suggestion list (plain input only, matching the design)
- Tag management page

## Decisions

### Where to fetch `GET /tags`

**Decision**: Fetch inside `RightPanel` with `useQuery(['tags'])`, same stale time as folders (60s).

**Rationale**: Hoisting to `App.tsx` would add coupling for a feature that only matters when the panel is open. TanStack Query deduplicates requests — if the panel is opened multiple times the cache is shared automatically. This mirrors how `folders` are fetched today (`RightPanel.tsx:61`).

**Alternative considered**: Fetch in `AppLayout` and pass down as a prop. Rejected — unnecessary prop threading for data that's only consumed by `RightPanel`.

---

### Tag state shape inside RightPanel

**Decision**: Local state is `{ id: string; name: string }[]`, initialized from `image.tags`.

**Rationale**: The PATCH payload needs IDs; the display needs names. Keeping both together avoids a lookup on every render and keeps the TagInput component simple (receives and emits the same shape).

**Alternative considered**: Two parallel arrays (`tagIds`, `tagNames`). Rejected — harder to keep in sync, more error-prone.

---

### When to fire PATCH

**Decision**: On every add or remove — immediate, no debounce.

**Rationale**: Consistent with how title/notes/source work in the existing panel (save on the event, not on a timer). Tags are committed explicitly by the user (Enter/comma/blur), so each commit is a deliberate action, not a keystroke stream.

**Alternative considered**: Debounced or batched save. Rejected — adds complexity and diverges from the existing panel's UX contract.

---

### PATCH tag set strategy

**Decision**: Send the full tag UUID array on every change (replace strategy).

**Rationale**: This matches what the backend implements (`ReplaceImageTags` deletes all then inserts). The FE always knows the current state from its local tag list, so computing the full set is trivial.

---

### Name-to-ID resolution

**Decision**: On commit, search the `allTags` cache by name (case-insensitive). If found, reuse the ID. If not found, call `POST /tags` → get the new ID → update the query cache → proceed to PATCH.

**Rationale**: Prevents duplicate tags silently. The user types a name; the system owns the ID.

## Risks / Trade-offs

- **Stale tag cache**: A tag created on another device won't appear in the local cache until the 60s stale window expires. Low risk for a single-user app.
- **Race on double-add**: If two tab instances add the same new tag name simultaneously, the second `POST /tags` will hit the 409 Conflict from the backend unique constraint. The handler should treat a 409 as "tag already exists" — re-fetch tags and resolve the ID.
- **Replace strategy on concurrent edits**: Two panels open for the same image (unlikely) could clobber each other's tag sets. Acceptable — no realistic multi-window scenario for this app.
