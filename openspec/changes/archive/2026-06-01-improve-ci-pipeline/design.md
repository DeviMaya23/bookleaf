## Context

Currently one workflow (`pr-go-tests.yml`) runs backend Go tests on every PR to main, unconditionally. There is no frontend CI at all. The workflow has no concurrency control, so multiple pushes to the same PR queue up redundant runs.

## Goals / Non-Goals

**Goals:**
- Path-filter BE and FE jobs so each only runs when relevant files change
- Add golangci-lint to backend CI
- Add ESLint, Vitest, and `npm run build` to frontend CI
- Cache npm dependencies for faster FE job execution
- Cancel stale in-progress runs when new commits are pushed to an open PR

**Non-Goals:**
- CD pipeline changes (deploy workflows untouched)
- Branch protection / required status checks configuration (out of scope for now)
- Code coverage reporting or thresholds
- Running CI on push to main (only PR triggers)

## Decisions

### Decision: Job-level path filtering via `dorny/paths-filter`, not workflow-level `paths:`

Using `paths:` on the workflow trigger causes the entire workflow to be skipped (neutral status) when no files match, rather than passing. This is problematic if required status checks are added later. Job-level filtering keeps the workflow always running with individual jobs conditionally skipped — a cleaner model for future branch protection adoption.

`dorny/paths-filter` outputs boolean flags from a single `detect-changes` job; downstream jobs read these via `needs.detect-changes.outputs.<flag>`.

**Alternative considered**: Separate workflow files per area (BE/FE), each with `paths:` trigger. Rejected because it fragments CI state across multiple workflow runs and makes the path-filtering trap harder to fix later.

### Decision: Unified `pr-ci.yml` replacing `pr-go-tests.yml`

One workflow file with three jobs (`detect-changes`, `backend-ci`, `frontend-ci`) is simpler to maintain than multiple files. All CI status is visible in one place on the PR.

### Decision: `golangci-lint` over standalone `go vet`

`golangci-lint` is a superset of `go vet` and runs errcheck, staticcheck, ineffassign, and others that catch real bugs. It has an official GHA action that handles its own caching. A minimal `backend/.golangci.yml` is added to pin enabled linters and avoid surprises from upstream default changes.

**Alternative considered**: `go vet` only. Rejected — it catches ~5 issues; golangci-lint is the Go ecosystem standard and costs nothing extra in CI time.

### Decision: `actions/setup-node` npm cache for frontend job

`actions/setup-node@v4` with `cache: 'npm'` and `cache-dependency-path: frontend/package-lock.json` caches the npm cache directory keyed to the lock file hash. `npm ci` then uses that cache. This is the idiomatic approach — caching `node_modules` directly is fragile across OS/Node version changes.

### Decision: FE job order — lint → test → build

ESLint runs first (fastest, catches simple issues), then Vitest, then `npm run build` (slowest, `tsc -b + vite build`). Fail-fast ordering means cheaper checks gate expensive ones.

## Risks / Trade-offs

- **First-run cache misses**: npm cache and golangci-lint cache are cold on first run; subsequent runs benefit. Low impact.
- **golangci-lint noise on existing code**: May surface pre-existing issues. Mitigation: configure `.golangci.yml` with a curated linter set; add `--new-from-rev` if needed.
- **`detect-changes` adds ~5s overhead**: Negligible. The job is a single `git diff` operation.
- **Workflow rename**: Any external reference to `PR Go Tests` by name (e.g., a future required status check) must use the new job name. Not a current concern since no branch protection is configured.

## Migration Plan

1. Delete `pr-go-tests.yml`
2. Create `pr-ci.yml` with all three jobs
3. Create `backend/.golangci.yml`
4. Verify on a test PR that all three jobs appear and path filtering works as expected
