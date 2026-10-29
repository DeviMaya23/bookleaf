# Conventions

> Agent-agnostic. Applies to all implementation agents.
> For stack and library choices, see `STACK.md`. For directory structure, see `ARCHITECTURE.md`. For testing, see `TESTING.md`.

---

## Linting

Do not suppress ESLint findings with `eslint-disable` comments — fix the underlying issue instead.

Exceptions (add a comment explaining why when using these):
- `react-hooks/exhaustive-deps`

---

## Frontend

### Component & Hook Granularity

When building a new component, split out a hook for each cohesive concern
(state + its effects + its handlers) rather than writing one component that
owns everything. A concern is "cohesive" if it could be described in one
sentence without mentioning the rest of the component — e.g. "the
pan/zoom/rotate engine," "the manual-reorder drag state," "the
delete/restore/hard-delete lifecycle."

- One hook per concern, with a colocated test file (`useThing.ts` +
  `useThing.test.ts`), written alongside the hook — not bolted on after the
  component is "done."
- Hooks return semantic mutators, not raw setters: `removeImage(id)`,
  `toggleFlip()`, `rotate()` — never `setImages`/`setRotation`. The hook owns
  the "how"; the caller expresses intent.
- Group related DOM handlers a hook exposes into one object meant to be
  spread (`dragHandlers`), matching the `attributes`/`listeners` convention
  dnd-kit already uses in this codebase — reuse existing conventions for hook
  return shapes where one applies.
- If a hook attaches effects to a DOM ref (`ResizeObserver`, native
  listeners), write its test around a small harness component + `render()`,
  not bare `renderHook` — the ref needs to exist before mount effects run.

### Matching Existing Shapes for Presentational Components

Before writing a new dialog, confirmation, or nav-row component, check
whether the feature (or a sibling feature) already has one of the same kind
and match its shape rather than inventing a new one:

- Confirmation dialogs: `{ item: T | null, onCancel, onConfirm }` for a
  nullable-target confirmation, or `{ open: boolean, onCancel, onConfirm }`
  for a plain yes/no with no target.
- A nav-row/list-item that owns its own state + mutation + dialog should be
  one self-contained file, mirroring how sibling rows in the same list are
  structured.
- The parent keeps owning orchestration state (which item is targeted, the
  mutation call); the extracted piece stays presentational unless the whole
  self-contained unit is what's being mirrored.

Don't introduce a new generic shared component (e.g. a `ConfirmDialog` in
`components/ui/`) to save a handful of lines on the first or second
occurrence — that's a new cross-feature abstraction and falls under
Decision Boundaries in `CLAUDE.md`. Two near-identical feature-owned components
are fine; reach for a shared base once a third occurrence appears or the
duplication is large.

### Avoiding Duplicated Logic Across Entry Points

Before copy-pasting an existing flow to wire up a new entry point (a new
upload trigger, a new chip-input, a new mutation pipeline), check whether it
should instead call into — or be factored alongside — the existing
implementation:

- Cross-cutting pipelines (a multi-step sequence like validate → transform →
  upload → finalize) used from more than one place belong in `lib/` as a
  single function, parameterized by the small bits that vary per caller. Keep
  validation separate from the pipeline so each caller's existing
  error-handling convention (throw vs. status flag vs. local state) still
  works without forcing one convention on everyone.
- Near-identical UI components (two inputs/widgets that would differ only in
  item type, copy, and one or two optional behaviors): build the shared
  generic implementation first, then thin per-case wrappers with their
  natural prop names — don't write two full implementations and plan to
  unify "later."
- The bar for "factor this out now" is duplicated logic with drift risk (a
  fix would need to land in N places) — not file size. A new feature file
  that's merely long but cohesive doesn't need splitting; a new feature that
  duplicates an existing multi-step flow does.

### Test Colocation

Each new hook or extracted component gets its test file written in the same
step it's created — not deferred to a "tests" pass at the end of the feature.
This keeps coverage attributable to the unit that owns the behavior, and
avoids a later split having to guess what the original intended to cover.

### State Management & Type Safety

**Local state that diverges from props isn't always a reset-on-prop-change anti-pattern**

`useState` + `useEffect` that re-syncs from a prop looks like the textbook
"reset state when prop changes" smell, but check whether the local state
exists to hold an **optimistic/in-flight value** that must temporarily diverge
from the prop (e.g. `useFieldAutosave`'s edit buffer, `useImageDetailsData`'s
`selectedFolders`, `useManualReorder`'s `orderedImages`). If so, the effect is
legitimate re-sync and should stay.

The real anti-pattern is local state that's **purely derived**, with no
divergence purpose (e.g. `RightPanel`'s old `tags` state) — fix with
`key`-based remount or `useMemo`, not a reset effect.

(Same applies to "ref mirrors latest state for stale-closure-safe access"
effects and DOM-measurement-driven reset effects — both are legitimate
external sync, not state-reset smells.)

**Type guards at library boundaries**

When a library erases your own data's shape to `any` (e.g. dnd-kit's
`Active.data.current`), and a producer/consumer rename could break things with
zero type signal, a small local type guard narrowing back to your own
discriminated union is worth it — no new dependency required.

---

## Backend

### Go Version and Dependencies

The `go` directive in `go.mod` is a deliberate choice, not a side effect. Go upgrades are only done when explicitly requested.

`go get` and `go mod tidy` silently raise the `go` (and add `toolchain`) line when a dependency declares a higher version. After any dependency change, check `git diff go.mod`. If `go`/`toolchain` changed:

1. Revert that change.
2. Look for an older version of the dependency that works with the current Go version (`go list -m -versions <module>`, then `go get <module>@<version>`).
3. If no compatible version exists, stop and tell the user which dependency requires which Go version. Do not bump it yourself.

---

## Browser Extension

All extension code MUST use `browser.*` from `webextension-polyfill`
instead of raw `chrome.*` APIs. If no polyfill equivalent exists,
flag it before proceeding.
