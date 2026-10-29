# Architecture

> Agent-agnostic. Applies to all implementation agents.

## Frontend

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

## Backend

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

## Browser Extension

The extension lives under `/extensions`. It targets Chrome MV3 by default and is built with Firefox compatibility in mind.
