## Context

SQL repositories currently have no test coverage. Unit tests with mocked databases were ruled out because they can't catch migration drift, constraint violations, or real query behavior. The solution is integration tests that run against a real Postgres instance managed by testcontainers-go.

The shared infrastructure lives in `internal/testutil/` and is consumed by each repository package's `main_test.go`. All migrations live in the existing `migrations/` folder and are applied once per test suite run.

## Goals / Non-Goals

**Goals:**
- Provide a zero-config way to spin up a Postgres container in tests
- Apply all existing migrations before any test runs
- Isolate each test's writes via transactions that roll back after the test
- Establish a consistent `TestMain` lifecycle pattern for all repository packages

**Non-Goals:**
- Replacing or modifying any production code paths
- Providing test helpers for service or handler layers (those use mocks)
- Parallelizing container startup across packages (each package gets its own container)
- Supporting databases other than Postgres

## Decisions

### testcontainers-go over docker-compose for container lifecycle

testcontainers-go manages container lifecycle programmatically inside the test binary. This means no external setup step, no leftover containers if a test crashes, and clean CI integration without a separate compose file.

Alternative considered: a shared `docker-compose.test.yml` with a pre-started Postgres. Rejected because it requires manual setup and teardown, and leaves state between runs.

### One container per package via `TestMain`, not per test

Spinning up a container is slow (~2–3s). Sharing one container across all tests in a package keeps the suite fast while transaction-level rollback provides per-test isolation.

Alternative considered: one container per test function. Rejected due to startup overhead multiplying with test count.

### Transaction rollback for test isolation

`NewTestTx` starts a transaction and registers `t.Cleanup(tx.Rollback)`. Each test writes into its own transaction and those writes are never committed. This avoids truncating tables between tests, which is slower and order-dependent.

Alternative considered: `TRUNCATE` tables in a `t.Cleanup`. Rejected because it requires knowing which tables to truncate and fails if a test creates new tables via DDL.

### Migrations applied via `golang-migrate` against the live container

The same migration tool used in production ensures the schema in tests matches production exactly. Migrations run once in `SetupPostgresContainer` before `m.Run()`.

## Risks / Trade-offs

- **Container startup latency** → Mitigated by sharing one container per package; CI pipelines with Docker available should see <5s overhead per package.
- **Flaky tests if Docker is unavailable** → Tests will fail fast with a clear error from testcontainers-go. Acceptable for a developer machine without Docker; CI always has Docker.
- **Parallel package execution may start multiple containers simultaneously** → Each container is independent and uses a dynamic port, so there is no port conflict. Memory usage scales with parallel package count but is bounded.

## Open Questions

- Should `SetupPostgresContainer` accept a custom migrations path, or always use `migrations/`? For now hardcoded to `migrations/` since there is only one migrations folder.
