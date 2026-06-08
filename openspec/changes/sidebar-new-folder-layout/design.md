## Context

`FolderSidebar.tsx` currently has two areas relevant to this change:

- The "FOLDERS" section label (`frontend/src/components/FolderSidebar.tsx:361-366`), a plain `<p>` with no adjacent controls.
- The footer block (`frontend/src/components/FolderSidebar.tsx:387-401`), which unconditionally renders a full-width "+ New folder" `<button>`, the folder filter input, and `ProfileMenu`.

Both the existing footer button and the new header icon button will open the same `FolderNameDialog` via the existing `setNewFolderOpen(true)` call — no change to the dialog, mutation, or `POST /folders` flow.

The folder list (`folders`, from `useQuery` at line 255, defaulting to `[]`) is already loaded in this component and is the natural source for determining "is the account empty."

## Goals / Non-Goals

**Goals:**
- Add a small icon button next to "FOLDERS" that opens the new-folder dialog, using the existing `ghost` / `icon-xs` (or `icon-sm`) `Button` convention (see `dialog.tsx` close button).
- Hide the footer "+ New folder" button once the user has at least one folder, so it serves purely as an onboarding prompt for empty accounts.

**Non-Goals:**
- No changes to the new-folder dialog, validation, mutation, or API contract.
- No changes to the folder filter input or `ProfileMenu` placement (raised separately, out of scope here).
- No change to "New subfolder via context menu" behavior.

## Decisions

**1. Emptiness check uses the existing `folders` array, not `visibleTree`.**
The footer button's visibility should reflect whether the *account* has any folders, not whether the current filter matches any. Using `folders.length === 0` (the unfiltered list already returned by the existing query) keeps the onboarding prompt tied to actual account state — typing into the filter shouldn't make the prompt reappear.

Alternative considered: deriving emptiness from `visibleTree`. Rejected — it would conflate "no folders exist" with "no folders match the filter," which would resurface the onboarding CTA in a confusing context.

**2. Header icon button reuses the `Button` UI component with `variant="ghost"` and `size="icon-xs"` (or `icon-sm`).**
This is the only icon-button convention already present in the codebase (the dialog close button). Reusing it avoids introducing a new visual pattern for icon-only controls. Final size choice (`icon-xs` 24px vs `icon-sm` 28px) is left to implementation/visual judgment given the sidebar's compact text scale — both are within the existing token set, so this isn't a new pattern decision.

**3. Conditional rendering is a simple ternary on `folders.length === 0`, not a separate empty-state component.**
The footer button already exists as a single JSX element; wrapping it in `{folders.length === 0 && (...)}` is the minimal change that satisfies the requirement without introducing a new abstraction (e.g., a dedicated `EmptyFolderPrompt` component), consistent with not over-engineering a one-element conditional.

## Risks / Trade-offs

- **[Risk]** A user with exactly one folder loses the larger, more discoverable CTA right after creating their first folder, which might feel abrupt. → **Mitigation**: This is the intended UX per the proposal — the header icon button is meant to be the steady-state affordance from that point forward; no further mitigation needed.
- **[Risk]** `folders.length === 0` flips on every successful creation/deletion, so the footer button could flicker in/out if a user deletes their last folder and recreates one. → **Mitigation**: This is the correct, intended behavior (empty ⇒ show prompt, non-empty ⇒ hide it); the existing `useQuery` refetch on mutation success already drives this naturally.

## Open Questions

None — sizing (`icon-xs` vs `icon-sm`) can be resolved visually during implementation without further design discussion.
