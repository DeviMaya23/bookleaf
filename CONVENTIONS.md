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

All extension code MUST use `browser.*` from `webextension-polyfill` 
instead of raw `chrome.*` APIs. If no polyfill equivalent exists, 
flag it before proceeding.