# Frontend review rubric

Stack: React + Vite + TypeScript, Tailwind, shadcn/ui on `@base-ui/react`, TanStack Query.

This file is the single source of truth for both the `fe-reviewer` subagent and the `/review-fe` command. Edit it here, not in those files.

## How to review

- **Flag, don't fix.** Report findings only. Never edit code files.
- **Plain language.** The reader is strong in Go and still learning React. For each finding say: what is wrong, why it hurts later, and (where one fits) a short Go analogy.
- **Cite `file:line`** for every finding.
- **Severity tiers**, in this order:
  1. **Bug**: wrong or fragile behavior now (stale data, broken a11y, crash on empty data).
  2. **Debt**: works today but makes future changes costly (duplication, misplaced code, redundant state).
  3. **Style**: preference only. Keep to at most three of these, or skip them.
- **Justify-or-flag.** For every `useEffect`, `useState`, `useMemo`, `useCallback`, and `useRef` in scope, state in one line why it exists. Flag any you cannot justify.
- **Don't pad.** If a section has no findings, say so in one line. Do not invent findings to look thorough.

## Modes

There are two modes. The caller says which. If the request names paths to audit, use audit mode. Otherwise use diff mode.

### Diff mode (default)

- Scope: changed lines under `frontend/`, plus the sibling components and hooks they touch, read only to judge consistency.
- Tag every finding **new** (introduced by this diff) or **pre-existing** (already in the code the diff touches or sits beside, not introduced by it). List new findings first.
- Do not review code the diff does not touch.

### Audit mode

- Scope: the named files or directories, read in full. No git diff.
- No new/pre-existing tag. Everything is existing code.
- **Group findings by pattern, not by file.** The same mistake in N places is one finding: name the pattern, give the count, and list every `file:line`. Order patterns by tier, then by number of occurrences.
- End with a **suggested fix order**: which pattern to fix first and why (real bugs first, then patterns that cause future bugs, then duplication in the most-used files). Skip Style findings entirely.
- In the hook and state audit, list only the unjustified ones, not every hook in scope.

## 1. Project rules (check mechanically)

Placement
- Feature-specific code lives in `features/<feature>/{components,hooks,lib}`. `app-shell/` is only for shell layout and cross-feature orchestration. `lib/` is for shared domain modules used by 4+ features.
- Promote to shared layers only when a second feature actually needs it. Flag speculative promotion.
- Feature directory names stay distinct from `lib/` domain names (`folder-sidebar`, `right-panel`, not `folders`, `images`).

Hooks and components
- One hook per cohesive concern (state + its effects + its handlers). A concern can be described in one sentence without mentioning the rest of the component. Flag components that own several concerns inline.
- Hooks return semantic mutators (`removeImage(id)`, `toggleFlip()`), never raw setters (`setImages`).
- Related DOM handlers a hook exposes are grouped in one spreadable object (like `dragHandlers`).
- New hooks, and extracted components that have behavior (state, effects, or handlers), have a colocated test file written in the same change. Purely presentational components (props in, markup out) do not need their own test. Hooks that attach to a DOM ref are tested through a small harness component, not bare `renderHook`.
- New dialogs, confirmations, and nav rows match the shape of existing siblings: `{ item: T | null, onCancel, onConfirm }` or `{ open, onCancel, onConfirm }`.
- No new shared abstraction (e.g. a generic `ConfirmDialog` in `components/ui/`) on the first or second occurrence. Two similar feature-owned components are fine. Reach for a shared base on the third.
- JSX in local variables (`const filtersContent = <div>...</div>`). A variable holding JSX hides what it depends on, because it reads state and props through closure rather than explicit props.
  - Rendered once: cosmetic. Suggest inlining it, as Style at most, never as Debt.
  - Rendered in more than one place: Debt, because the name hides duplicated rendering. Extract a component that takes the data as props (pass one object, not field by field) and keep each call site's placement, visibility classes, and show condition outside it. Do not suggest pasting the markup into each location.
  - If the multiple renders are both mounted at once with one hidden by CSS, note that both instances mount (duplicate IDs, double effects, tests matching two elements). That is a layout question, not a reason to skip the extraction.

Duplication
- A multi-step flow (validate, transform, upload, finalize) used from more than one place belongs in `lib/` as one function. Flag copy-pasted flows.
- Near-identical components should share a generic implementation with thin wrappers. The bar is drift risk (a fix would need to land in N places), not file length.

State
- Local state that re-syncs from a prop via `useEffect` is **legitimate** if it holds an optimistic or in-flight value that must diverge from the prop. It is a **smell** if the state is purely derived. The fix there is `useMemo` or a `key` remount.
- Ref-mirrors-latest-state effects and DOM-measurement reset effects are legitimate.
- Where a library erases your data's shape to `any` (e.g. dnd-kit `Active.data.current`), prefer a small local type guard back to the project's union.

Libraries and lint
- No imports from `@radix-ui/*`. Base UI is the primitive layer; check `@base-ui` docs, not Radix.
- Use `onClick`, not `onSelect`, on `ContextMenuItem`.
- No `eslint-disable` comments. The only exception is `react-hooks/exhaustive-deps`, and it needs a comment explaining why.

## 2. Generic checks (not yet written down as project rules)

Responsive layout
- Desktop vs mobile differences should be one component with Tailwind breakpoints (`sm:`, `md:`, ...). Flag `useMediaQuery`, `window.innerWidth`, or `isMobile ? <A /> : <B />` unless there is a stated reason. Duplicate trees mean every change has to be made twice.
- Flag layout that only works at one width: fixed pixel widths, missing `min-w-0` on flex children, overflow with no scroll container, hover-only interactions with no touch alternative.

Dialogs and modals
- Use the Base UI / shadcn dialog primitive. Flag hand-rolled overlays, or code that bypasses the primitive's focus trap, Escape handling, or scroll lock.
- Check focus returns to the trigger on close, and that the modal has an accessible title.
- Check state resets correctly when the modal reopens with different data.

Server state (TanStack Query)
- Data fetching goes through `useQuery` / `useMutation`. Flag `useEffect` + `fetch` + `useState`.
- Flag server data copied into `useState` (except the optimistic-buffer case above).
- Flag `data = []` style defaults on query results that are then used to decide "empty". Loading and empty are different states; gate on `isPending` / `isError` first.
- Flag local `isSubmitting`-style state that duplicates a mutation's `isPending`.
- Mutations must invalidate or update every query whose data they change. Flag a mutation with no `onSuccess` invalidation or cache update.
- Query keys must be consistent and include every variable the query depends on. Flag keys built ad hoc inline in several places.
- Flag `enabled` flags or `select` functions that silently hide errors.

Loading, error, empty
- Every data-driven view handles loading, error, and empty. Flag views that render `undefined` data or only the happy path.

TypeScript hygiene
- Flag `any`, `as` casts (especially `as unknown as`), and non-null `!`. Each needs a reason or a type guard.
- Flag props typed wider than they need to be, and optional props that are always passed.

React fundamentals
- Flag `useEffect` that does not synchronize with something external (derived values, event responses, prop-to-state copying).
- Flag two pieces of state that must always agree, or state that duplicates props or query data.
- Flag array `key` props that use index on reorderable or deletable lists.
- Flag inline object/array/function props passed to memoized children, or `useMemo`/`useCallback` with no measured need.
- Flag components over roughly 200 lines that mix fetching, form state, and layout. Length alone is not a finding, so cite the mixed concerns.

## 3. Running rules

Frontend rules live in the `## Frontend` section of `CONVENTIONS.md` (also check `CLAUDE.md` if it has frontend rules). Treat every rule there as part of section 1. When the reader accepts a finding as a standing rule, a line is added to that section. Rules there override this file if they conflict.

## Output format

### Diff mode

1. **Summary**: two or three sentences on the overall health of the change.
2. **Look at these first**: the top three findings, whatever their tier.
3. **Bug**, **Debt**, **Style**: findings grouped by tier. Each is `file:line`, new or pre-existing, what, why, Go analogy where useful.
4. **Hook and state audit**: one line per `useEffect` / `useState` / `useMemo` / `useCallback` / `useRef` touched, saying why it exists, with unjustified ones marked.
5. **Possible new rules**: any pattern seen twice in this diff that might deserve a line in `CONVENTIONS.md`. Suggest only. Do not write it.

### Audit mode

1. **Summary**: scope reviewed (paths, file count) and two or three sentences on overall health.
2. **Patterns**: one entry per pattern, ordered by tier then occurrences. Each has the tier, a name, the count, what and why (with Go analogy where useful), and every `file:line`.
3. **Unjustified hooks and state**: only the ones with no good reason to exist.
4. **Suggested fix order**: numbered, with a one-line reason for each step.
5. **Possible new rules**: patterns that recur enough to deserve a line in `CONVENTIONS.md`. Suggest only.
