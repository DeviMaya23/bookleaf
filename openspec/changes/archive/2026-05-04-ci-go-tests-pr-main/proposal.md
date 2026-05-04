## Why

Pull requests to `main` currently have no enforced automated backend test gate. Adding GitHub Actions coverage for `go test ./...` (including integration tests that use test containers) catches regressions before merge and makes CI expectations consistent for contributors.

## What Changes

- Add a GitHub Actions workflow that runs on pull requests targeting `main`.
- Configure the job to run backend Go tests with container support required by integration tests.
- Ensure CI environment provides Docker/test-container prerequisites so integration tests execute in workflow runs.
- Document CI behavior and any required test environment assumptions in repository configuration/docs.

## Capabilities

### New Capabilities
- `ci-pr-go-tests`: automated PR-to-main workflow that runs backend Go test suite, including integration tests that require test containers.

### Modified Capabilities
- `sql-repo-test-infra`: extend requirements so SQL repository integration tests are expected to run in CI (not only local) with container runtime availability.

## Impact

- **Affected code/config:** `.github/workflows/*`, backend test invocation configuration, and related test infra docs/config
- **APIs:** No application API changes
- **Dependencies/systems:** GitHub Actions runners with Docker availability for test containers
