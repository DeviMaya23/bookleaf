# Conventions

> Agent-agnostic. Applies to all implementation agents (Claude Code, Copilot, Cursor, etc).
> For project information, refer to PROJECT.md.
> For agent process and workflow, refer to CLAUDE.md.

---

## Stack

### Backend
- Language: Go
- Framework: Echo
- ORM: GORM
- Database: PostgreSQL
- Auth: Kinde
- Storage: Cloudflare R2
- Logging: Zap
- Observability: OpenTelemetry

### Frontend
- Framework: React + Vite + TypeScript
- Styling: Tailwind CSS
- UI components: shadcn/ui via `@/components/ui/*`
- UI primitives: `@base-ui/react`

---

## UI Primitives

- UI components use `@base-ui/react`, NOT Radix UI
- `asChild` is a Radix UI pattern — it does not exist in this codebase and will cause a TypeScript error
- To render a trigger or wrapper with button styling, apply `buttonVariants()` via `className` directly on the component instead of wrapping it in `<Button asChild>`
- Never import directly from `@radix-ui/*`
- Application code imports only from `@/components/ui/*`. Direct `@base-ui/react` imports stay inside `src/components/ui/` only.

## Linting

Do not suppress ESLint findings with `eslint-disable` comments — fix the underlying issue instead.

Exceptions (add a comment explaining why when using these):
- `react-hooks/exhaustive-deps`

---

## Library Usage

### Frontend
Before self-implementing any non-trivial logic — layout algorithms, data
transformation pipelines, anything with meaningful edge cases or more than
~20 lines — you must:

1. Flag that you are about to self-implement something
2. Note any well-known libraries that solve the same problem
3. State your reasoning for or against each option
4. Wait for a decision before proceeding

This rule exists because the React/JS ecosystem has strong conventions
that are not always obvious. Do not assume self-implementation is preferred.

### Backend (Go)
Flag non-trivial self-implementations and surface library alternatives,
but a decision is not required before proceeding — note the trade-off
inline and continue unless the implementation touches a shared interface
or introduces a new architectural pattern (see Decision Boundaries in CLAUDE.md).

---

## Frontend Architecture

### Directory Structure

```
frontend/src/
  features/          ← feature-scoped code, one folder per UI area
    <feature>/
      components/    ← components owned by this feature
      hooks/         ← hooks owned by this feature
      lib/           ← pure/feature-local helpers (no other feature imports them)
  app-shell/         ← the top-level layout (AppLayout) and cross-cutting
                       orchestration (drag-and-drop, view routing)
    lib/             ← shell-local helpers (e.g. dragHandlers)
  lib/               ← shared domain layer: types + API wrappers used across
                       multiple features (images, folders, tags, thumbnail,
                       view, api, utils, browser, fracdex)
  components/ui/     ← shared design-system primitives (shadcn/Base UI)
  hooks/             ← generic hooks with no feature-specific dependencies
  pages/             ← thin route entry points
  assets/            ← static assets
```

Current features: `gallery`, `viewer`, `folder-sidebar`, `right-panel`,
`upload`, `auth`.

### Where new code goes

**`features/<feature>/`** — default for anything specific to one UI area. New
components, hooks, and helpers for that feature live here, not in top-level
`components/`/`hooks/`/`lib/`.

**`app-shell/`** — only the shell layout and logic that genuinely coordinates
across features (drag-and-drop orchestration, view routing). Do not put
feature work here.

**`lib/`** — shared domain modules consumed by 4+ features. Keep these
feature-agnostic; do not relocate a shared module into a single feature. A
helper used by only one feature belongs in that feature's `lib/`, not here.

**`components/ui/`, `hooks/`, `pages/`** — stay top-level: design-system
primitives, dependency-free generic hooks, and thin route entry points
respectively.

### Naming

Feature directories whose UI concern maps to a shared domain module use a
distinct name (`folder-sidebar`/`right-panel`, **not** `folders`/`images`), so
the feature directory and the `lib/` domain module stay separately greppable.

### Promotion (YAGNI)

A feature-local component/hook/helper is promoted to the shared layer
(`components/`, `hooks/`, `lib/`) only once a second feature actually needs it.
Do not pre-promote speculatively.

---

## Frontend Development Practices

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
CLAUDE.md's Decision Boundaries. Two near-identical feature-owned components
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

---

## Frontend State Management & Type Safety

### Local state that diverges from props isn't always a reset-on-prop-change anti-pattern

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

### Type guards at library boundaries

When a library erases your own data's shape to `any` (e.g. dnd-kit's
`Active.data.current`), and a producer/consumer rename could break things with
zero type signal, a small local type guard narrowing back to your own
discriminated union is worth it — no new dependency required.

---

## Backend Architecture

### Directory Structure

```
backend/
  internal/
    domain/          ← entities and value objects; no dependencies on other layers
    usecase/         ← business logic; defines interfaces for everything it depends on
    repository/      ← database access; implements interfaces defined in usecase/
    handler/         ← HTTP delivery layer
      middleware/    ← HTTP-specific middleware
    worker/          ← background job workers (River queue)
    platform/        ← cross-cutting infrastructure; startup concerns only
      config/        ← environment/config parsing
      observability/ ← OpenTelemetry, logging, metrics setup
    testutil/        ← integration test helpers
  pkg/               ← generic utilities with no app-specific knowledge
```

### Interfaces

Interfaces are defined by their consumer, in the consumer's package.

The usecase layer defines every interface it depends on — repositories, external
services, anything it calls. Implementing packages satisfy the interface without
importing from `usecase/`.

**File placement inside `usecase/`:**
- Inline in the usecase file if: ≤ 2 methods AND used only by that file
- Separate file named `<resource>_<dependency>.go` if: > 2 methods OR shared across usecases

**Visibility:** all interfaces are exported (public), regardless of placement. The placement rule already signals scope; visibility does not need to carry a second meaning.

### Package placement

**`internal/`** — default for all app code. Any package that imports from
`internal/domain`, `internal/platform/config`, uses app credentials, or is
otherwise coupled to this codebase stays here.

**`pkg/`** — zero app-specific knowledge; could be dropped into any Go project
unchanged. Requires a deliberate decision to promote. Do not place packages here
speculatively.

**`handler/middleware/`** — HTTP middleware only. If other transports are added
(e.g. gRPC), their middleware lives under that transport's directory, not at the
top level of `internal/`.

### Exceptions

`repository/` may contain ORM infrastructure files (e.g. a custom GORM logger)
alongside repository implementations. These do not need to move to `platform/`.

---

## Backend Testing

### Philosophy

Test usecase behaviour, not structure. A test that only verifies a method was called, or that an error passed through unchanged, tests implementation — not correctness. Integration tests cover real outcomes; unit tests cover logic that lives exclusively in the usecase.

### Layer coverage

| Layer | Test type |
|---|---|
| `usecase/` | Unit tests |
| `handler/` | Unit tests |
| `repository/` | Integration tests only — no unit tests |

### What to test in usecases

**Happy path** — only if the usecase adds assembly or transformation. If the function delegates to a repo and returns the result unchanged, the happy path is covered by integration tests.

Worth testing:
- Assembling a composite result from multiple sources
- Computing a derived value (path construction, filename, cursor)
- Conditional output based on branching logic

**Exception:** external services not backed by integration tests (e.g. `StorageService`, `VisionService`, `ThumbnailService`). Happy paths that assemble results from these are worth unit testing even without transformation, because there is no integration fallback.

**Error path** — only if the usecase adds behaviour to the error:
- Wrapping: `fmt.Errorf("context: %w", err)`
- Mapping to a domain sentinel: `ErrInvalidFolderName`, `ErrDuplicateTagName`
- Side effects or conditional logic triggered on the error path

Do not write error path tests for pure pass-through (`if err != nil { return nil, err }` with no other logic).

### What not to test

- Pure delegation — function calls one dependency and returns the result unchanged
- Error pass-through — `return nil, err` with no wrapping, mapping, or side effects
- Logging side effects — not verifiable in unit tests

### Test doubles

**Use a fake** when the test requires setting up pre-state in a dependency AND asserting post-state. The signal is: the function reads existing state to decide what to write.

```go
// AcceptSuggestion reads whether a folder exists → branches → writes image folder assignment
// Fake starts with: folder "Nature" exists or doesn't
// Fake ends with:   image is in folder "Nature"
// Assert the outcome, not the call graph
```

**Use a value-return spy** for everything else — single-step operations, external services, and cases where what got called IS the verifiable behaviour (e.g. confirming a no-op skipped a write).

Test doubles implement only the narrow interface defined by the consuming usecase — not the full repository interface. If the usecase defines `type imageCounter interface { CountByFolderID(...) }`, the test double implements only that one method.

### Table-driven tests

Use TDT when all of the following are true:
- 3 or more scenarios
- All cases share identical setup structure
- Cases differ only in inputs, outputs, or expected errors

Use individual `t.Run` blocks when cases need different setup, or when there are fewer than 3 scenarios. Do not stretch TDT to fit cases with varying mock configurations.

### Assertion quality

- Assert the result, not just the error
- Failure scenarios must assert the specific error type or message — not just that an error occurred
- Use `require.ErrorIs` for sentinel errors, `require.ErrorContains` for message matching

### Repository integration tests

Repository tests run against a real database via testcontainers. They verify the **database contract** — what each method promises about its interaction with the schema.

**Always test:**

- **Query correctness** — the right records are returned with the right fields, including preloaded associations where applicable
- **Ownership / user isolation** — a record belonging to user A is not accessible to user B. This is a security property and the most commonly missing test. Methods that scope by `userID` must have a wrong-user case asserting `gorm.ErrRecordNotFound`
- **Soft delete semantics** — soft-deleted records are excluded from standard queries and visible only in explicit trash/deleted queries
- **Constraint behaviour** — unique constraints return the expected error; FK violations are caught
- **Cascade behaviour** — `DeleteWithCascade` and similar operations verify that related rows are actually affected

**Do not test:**

- Infrastructure failures by closing the DB connection — this tests the driver, not the contract. Drop any test that forces an error by calling `sqlDB.Close()`

**Assertion style** — use `require` / `assert` from testify throughout. Do not use bare `if err != nil { t.Fatalf }` checks.

### Test double placement

| Double | Location | Reason |
|---|---|---|
| Fakes (repository interfaces) | `usecase/fakes_test.go` | shared across test files in the package; real logic should not be duplicated |
| Spies (external services) | inline in `*_test.go` | small, vary per scenario, no sharing benefit |
| Spies (handler usecase mocks) | inline in `handler/*_test.go` | per-handler interfaces, nothing to share across files |

### External services

External services (e.g. `StorageService`, `VisionService`) are thin adapters over third-party systems. They delegate directly to an SDK or HTTP client with no domain logic.

- No unit tests on the implementation — correctness depends on the external system, not on code in this repo
- When used as a dependency in usecase tests, always use a value-return spy
- Integration tests only if the adapter contains non-trivial logic worth verifying against the real system

### Handler tests

The handler's job is: parse request → extract auth → call usecase → map errors → serialize response. Tests verify that plumbing is correct.

**Always test:**
- Happy path — assert both the HTTP status code and the response body shape. The handler assembles the HTTP response; that assembly is handler logic even when the usecase result passes through unchanged.
- Each distinct error mapping — `ErrInvalidFolderName → 400`, `gorm.ErrRecordNotFound → 404`, generic error → 500. Each unique mapping is a test case.
- Binding failures — malformed JSON body, invalid UUID path param.

**Request parsing** — handler tests may assert what was passed down to the usecase spy (e.g. `lastDescription`, `lastUpdateParams`) when the interesting logic is in how the handler parses or extracts the request. This is behavioral for the handler layer.

**Test doubles** — always value-return spies for usecase dependencies. Fakes are not used at the handler layer.

**No equivalent of "pure delegation, drop the test"** — even a handler that returns 204 with no body has error mapping worth testing.

---

## Browser Extension

All extension code MUST use `browser.*` from `webextension-polyfill` 
instead of raw `chrome.*` APIs. If no polyfill equivalent exists, 
flag it before proceeding.