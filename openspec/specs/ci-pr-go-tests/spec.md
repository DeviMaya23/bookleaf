# ci-pr-go-tests Specification

## Purpose
TBD - created by archiving change ci-go-tests-pr-main. Update Purpose after archive.
## Requirements
### Requirement: Pull Request Test Workflow
The repository SHALL include a GitHub Actions workflow file at `.github/workflows/pr-ci.yml` (replacing `pr-go-tests.yml`) that runs backend and frontend CI jobs for pull requests targeting the `main` branch.

#### Scenario: Pull request to main triggers workflow
- **WHEN** a pull request is opened, synchronized, or reopened with `main` as the base branch
- **THEN** the workflow starts automatically
- **AND** it executes the `detect-changes` job followed by conditionally gated backend and frontend jobs

### Requirement: Backend Go Test Execution
The workflow SHALL run `go test ./...` from the `backend/` directory, after linting passes.

#### Scenario: Test command runs from backend after lint
- **WHEN** the backend CI job executes and linting succeeds
- **THEN** it sets `backend/` as the working directory
- **AND** runs `go test ./...`

### Requirement: Integration Test Container Support
The workflow SHALL run on a runner environment that supports Docker so integration tests using testcontainers-go can start required containers.

#### Scenario: Integration tests requiring containers run in CI
- **WHEN** `go test ./...` executes in the workflow
- **THEN** integration tests that start Postgres test containers can run
- **AND** the workflow fails if those tests fail

### Requirement: Backend Path Filtering
The backend CI job SHALL only execute when files under the `backend/` directory have changed, determined by a `detect-changes` job using `dorny/paths-filter`.

#### Scenario: Path filter detects backend changes
- **WHEN** the `detect-changes` job runs
- **THEN** it outputs a boolean `backend` flag based on whether any `backend/**` files changed
- **AND** the backend CI job runs only if `backend` is `true`

#### Scenario: PR with no backend changes skips backend job
- **WHEN** a pull request to `main` is opened or updated with no changes under `backend/**`
- **THEN** the backend CI job is skipped

### Requirement: Go Linting with golangci-lint
The backend CI job SHALL run `golangci-lint` before running tests, using the official `golangci/golangci-lint-action` and a `backend/.golangci.yml` configuration file.

#### Scenario: Linting passes with no violations
- **WHEN** `golangci-lint` runs against the backend source
- **THEN** the step succeeds and the job continues to run tests

#### Scenario: Linting fails on a violation
- **WHEN** `golangci-lint` detects a violation (e.g., unchecked error, unused variable)
- **THEN** the step fails and the backend CI job is marked failed without running tests

### Requirement: Concurrency Group
The CI workflow SHALL define a concurrency group scoped to the workflow name and PR ref, with `cancel-in-progress: true`, so that a new push to an open PR cancels any in-progress run for that PR.

#### Scenario: New commit cancels in-progress CI run
- **WHEN** a new commit is pushed to a PR that already has a CI run in progress
- **THEN** the in-progress run is cancelled
- **AND** a new run starts for the latest commit

