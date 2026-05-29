## Context

`TagInput` is currently API-blind — it receives `tags` and calls `onChange`, with no knowledge of existing user tags. `RightPanel` already holds `allTags` (fetched via `useQuery(['tags'])`). This change threads suggestion data from `RightPanel` down to `TagInput` as a prop, then renders a dropdown inside `TagInput` while the user is typing.

## Goals / Non-Goals

**Goals:**
- Show a filtered suggestion list while the user is typing in `TagInput`
- Support mouse click and keyboard (↑↓ to navigate, Enter to select, Escape to dismiss)
- Exclude already-applied tags from suggestions
- Zero new API calls — filter is computed from the cached `allTags`

**Non-Goals:**
- Fuzzy matching — prefix/substring match is sufficient
- A standalone tag management UI (rename, delete)
- Suggestions when the input is empty (only show on active typing)

## Decisions

### Prop threading vs. context

**Decision**: Pass `suggestions: { id: string; name: string }[]` as a prop to `TagInput`. `RightPanel` computes the filtered list before passing.

**Rationale**: `TagInput` stays API-blind and testable in isolation. The filtering logic (`allTags` minus `tags` applied to image, filtered by current input) lives in `RightPanel` where both data sources are already available. A React context would be overkill for a single consumer.

**Alternative considered**: Pass the full `allTags` to `TagInput` and let it filter internally. Rejected — it blurs the component's responsibility boundary and requires `TagInput` to know about applied tags to exclude them.

---

### Dropdown positioning

**Decision**: Render the dropdown as an absolutely positioned `div` below the tag input container, using `position: absolute` within a `position: relative` wrapper on the container div.

**Rationale**: No extra dependency (no Floating UI / Popper). The right panel is a fixed-width sidebar so overflow is not a concern. A portal would add complexity for minimal gain.

---

### When to show / hide

**Decision**: Show the dropdown when `val.trim().length > 0` and `suggestions.length > 0`. Hide on Escape, on suggestion select, on blur (with a short delay to allow click to register), and when input is cleared.

**Rationale**: Showing suggestions on an empty input creates visual noise — the user hasn't expressed intent yet. The blur delay (e.g., `setTimeout 150ms`) is the standard pattern for allowing `mousedown` on a dropdown item to fire before `blur` hides it.

---

### Keyboard navigation state

**Decision**: Track a `selectedIndex` state (number, -1 = none selected). Arrow keys increment/decrement within the suggestions list. Enter at `selectedIndex >= 0` commits the highlighted suggestion. Enter at `selectedIndex === -1` commits the raw typed value (existing Enter behaviour preserved).

**Rationale**: Keeps the existing Enter-to-commit behaviour intact when the user ignores the dropdown. Natural fallthrough.

## Risks / Trade-offs

- **Blur race with mousedown**: Dismissing the dropdown on blur before the click registers is a classic problem. Mitigated with a 150ms timeout before hiding, which is sufficient for a click event to fire.
- **Stale suggestions prop**: If `allTags` cache is stale (new tag created on another device), suggestions won't include it. Acceptable — same limitation already exists for the core tagging flow.
