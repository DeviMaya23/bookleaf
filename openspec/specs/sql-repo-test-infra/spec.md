## Purpose

Provide test infrastructure and helpers for SQL repository integration tests using testcontainers-go, enabling reliable database testing in both local and CI environments.

## Requirements

### Requirement: Postgres container setup
The system SHALL provide a `SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error)` helper in `internal/testutil/` that starts a Postgres container via testcontainers-go and runs all migrations from the `migrations/` folder before returning.

#### Scenario: Successful container startup
- **WHEN** `SetupPostgresContainer` is called with a valid context
- **THEN** it returns a running `*PostgresContainer` with all migrations applied and no error

#### Scenario: Container startup fails
- **WHEN** Docker is unavailable or the container fails to start
- **THEN** it returns a non-nil error and a nil `*PostgresContainer`

### Requirement: Test database connection
The system SHALL provide a `NewTestDB(container *PostgresContainer) *gorm.DB` helper that returns a `*gorm.DB` connected to the given container, safe to share across all tests within a package.

#### Scenario: Successful DB connection
- **WHEN** `NewTestDB` is called with a running `*PostgresContainer`
- **THEN** it returns a non-nil `*gorm.DB` that can execute queries against the container

### Requirement: Transaction-scoped test handle
The system SHALL provide a `NewTestTx(t *testing.T, db *gorm.DB) *gorm.DB` helper that begins a transaction and registers a rollback via `t.Cleanup`, so writes made during the test are never persisted.

#### Scenario: Writes rolled back after test
- **WHEN** a test uses `NewTestTx` to insert a row and the test finishes
- **THEN** the row is absent from the database when the next test runs

#### Scenario: Rollback on test failure
- **WHEN** a test using `NewTestTx` fails before completing
- **THEN** the transaction is still rolled back via `t.Cleanup`

### Requirement: TestMain lifecycle per repository package
Each SQL repository package SHALL include a `main_test.go` that uses `TestMain(m *testing.M)` to: spin up a `PostgresContainer` before `m.Run()`, run all tests, and terminate the container after `m.Run()` returns.

#### Scenario: Container shared across tests in a package
- **WHEN** a repository package's tests run
- **THEN** all `TestXxx` functions share a single container started in `TestMain`

#### Scenario: Container terminated after suite
- **WHEN** all tests in a package finish (pass or fail)
- **THEN** the Postgres container is stopped and removed

### Requirement: No unit tests for SQL repositories
SQL repository packages SHALL NOT contain unit tests with mocked databases. All repository test coverage SHALL be provided exclusively through integration tests using the `TestMain` pattern.

#### Scenario: Integration test is the only test form for repositories
- **WHEN** a new SQL repository is added
- **THEN** its test file uses `NewTestTx` against a real container, not a mock `*gorm.DB`
