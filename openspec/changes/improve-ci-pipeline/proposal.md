## Why

The existing CI pipeline runs backend Go tests on every PR to main with no path filtering and no frontend checks, making it both slower than necessary (runs when only FE changed) and incomplete (TS errors and FE test failures go undetected). The goal is to make CI accurate — running the right checks for the right changes.

## What Changes

- Replace `pr-go-tests.yml` with a unified `pr-ci.yml` workflow
- Add a `detect-changes` job using `dorny/paths-filter` to gate BE and FE jobs by path
- Backend job: add `golangci-lint` before `go test`, only runs when `backend/**` changed
- Frontend job: run ESLint, Vitest, and `npm run build` (catches TS errors), only runs when `frontend/**` changed; use `actions/setup-node` with npm caching
- Add concurrency group to cancel in-progress runs on the same PR when new commits are pushed
- Add `backend/.golangci.yml` to configure linter selection

## Capabilities

### New Capabilities

- `ci-pr-fe-checks`: Frontend CI checks on PR to main — ESLint, Vitest unit tests, and TypeScript build validation; path-filtered to `frontend/**` changes only

### Modified Capabilities

- `ci-pr-go-tests`: Add path filtering (only runs on `backend/**` changes), add `golangci-lint` step before `go test`, rename workflow file from `pr-go-tests.yml` to `pr-ci.yml`

## Impact

- `.github/workflows/pr-go-tests.yml` — deleted (replaced)
- `.github/workflows/pr-ci.yml` — new unified workflow file
- `backend/.golangci.yml` — new linter configuration file
