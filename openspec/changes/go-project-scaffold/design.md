## Context

Greenfield Go project with no existing code. The scaffold needs to be correct and unblocking — the goal is to establish the clean architecture directory layout and a runnable entry point so all subsequent domain, repository, and handler work has a clear home.

## Goals / Non-Goals

**Goals:**
- Runnable Echo server with a health-check route
- Clean architecture directories (`domain/`, `usecase/`, `repository/`, `handler/`) tracked in git
- `go.mod` initialized with the correct module path
- Port configurable via environment variable

**Non-Goals:**
- Database connection (belongs with the first repository layer task)
- Clerk auth middleware (belongs with the auth change)
- Config file or structured logging (premature for a scaffold)
- `pkg/` directory (create only when there is genuinely shared code)

## Decisions

### Decision 1: Module path

Use `github.com/devi/bookleaf` — matches git user and project name, standard Go convention.

### Decision 2: Port via environment variable

Read port from `PORT` env var, default to `8080`. No config struct or file yet — that complexity belongs in a later change when there are multiple values to manage.

### Decision 3: No `pkg/` directory

Only create it when there is something genuinely shared between packages. An empty `pkg/` directory adds noise without value.

## Risks / Trade-offs

- Module path assumes the repo stays at `github.com/devi/bookleaf`. If it moves, a global find-and-replace on the module path is needed → low risk for an early-stage project.
