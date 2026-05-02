## Why

SQL repositories have no test coverage because unit tests with mocked databases provide false confidence — they can't catch query errors, constraint violations, or migration drift. We need real-database integration tests, and a shared infrastructure layer will make each repository package's test setup consistent and low-boilerplate.

## What Changes

- Add `internal/testutil/` package with three helpers: `SetupPostgresContainer`, `NewTestDB`, and `NewTestTx`
- Introduce a `TestMain`-based lifecycle pattern for all SQL repository test packages
- Establish a project-wide convention: no unit tests for SQL repositories, integration tests only

## Capabilities

### New Capabilities

- `sql-repo-test-infra`: Shared test helpers in `internal/testutil/` that spin up a real Postgres container via testcontainers-go, run migrations, and provide transaction-scoped DB handles for isolated test runs

### Modified Capabilities

## Impact

- Adds `github.com/testcontainers/testcontainers-go` as a test dependency
- All future SQL repository packages must include a `main_test.go` using the `TestMain` pattern defined here
- No changes to production code, APIs, or existing specs
