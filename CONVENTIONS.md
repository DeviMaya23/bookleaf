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

## Browser Extension

All extension code MUST use `browser.*` from `webextension-polyfill` 
instead of raw `chrome.*` APIs. If no polyfill equivalent exists, 
flag it before proceeding.