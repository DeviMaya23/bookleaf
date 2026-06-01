# ci-pr-fe-checks Specification

## Purpose
TBD

## Requirements

### Requirement: Frontend CI Workflow Trigger
The CI workflow SHALL include a frontend job that runs on pull requests targeting the `main` branch, triggered on opened, synchronize, and reopened events.

#### Scenario: PR with frontend changes triggers frontend job
- **WHEN** a pull request to `main` is opened or updated with changes under `frontend/**`
- **THEN** the frontend CI job runs

#### Scenario: PR with no frontend changes skips frontend job
- **WHEN** a pull request to `main` is opened or updated with no changes under `frontend/**`
- **THEN** the frontend CI job is skipped

### Requirement: Frontend Path Filtering
The frontend CI job SHALL only execute when files under the `frontend/` directory have changed, determined by a `detect-changes` job using `dorny/paths-filter`.

#### Scenario: Path filter detects frontend changes
- **WHEN** the `detect-changes` job runs
- **THEN** it outputs a boolean `frontend` flag based on whether any `frontend/**` files changed
- **AND** the frontend CI job runs only if `frontend` is `true`

### Requirement: Node.js Setup with npm Cache
The frontend CI job SHALL set up Node.js using `actions/setup-node@v4` with npm caching keyed to `frontend/package-lock.json`.

#### Scenario: npm cache hit on repeated runs
- **WHEN** `frontend/package-lock.json` has not changed since the last run
- **THEN** the npm cache is restored and `npm ci` completes faster than a cold install

#### Scenario: npm cache miss on lock file change
- **WHEN** `frontend/package-lock.json` has changed
- **THEN** `npm ci` performs a full install and the cache is updated

### Requirement: Frontend Linting
The frontend CI job SHALL run ESLint via `npm run lint` from the `frontend/` directory.

#### Scenario: Lint passes with no violations
- **WHEN** `npm run lint` executes
- **THEN** the step succeeds and the job continues

#### Scenario: Lint fails on ESLint violation
- **WHEN** `npm run lint` finds a rule violation
- **THEN** the step fails and the job is marked failed

### Requirement: Frontend Unit Tests
The frontend CI job SHALL run Vitest unit tests via `npm test` from the `frontend/` directory.

#### Scenario: All unit tests pass
- **WHEN** `npm test` executes
- **THEN** all `.test.tsx` / `.test.ts` files run and the step succeeds

#### Scenario: A unit test fails
- **WHEN** one or more Vitest tests fail
- **THEN** the step fails and the job is marked failed

### Requirement: Frontend TypeScript Build Validation
The frontend CI job SHALL run `npm run build` from the `frontend/` directory to validate that there are no TypeScript compilation errors and the production bundle can be created.

#### Scenario: Build succeeds with no TS errors
- **WHEN** `npm run build` executes (`tsc -b && vite build`)
- **THEN** the step succeeds

#### Scenario: Build fails on TypeScript error
- **WHEN** `npm run build` encounters a TypeScript type error
- **THEN** the step fails and the job is marked failed
