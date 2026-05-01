## 1. Dependencies

- [ ] 1.1 Add `github.com/testcontainers/testcontainers-go` and `github.com/testcontainers/testcontainers-go/modules/postgres` to `go.mod` / `go.sum`

## 2. testutil Package

- [ ] 2.1 Create `internal/testutil/` package
- [ ] 2.2 Implement `PostgresContainer` struct (wraps testcontainers container + connection string)
- [ ] 2.3 Implement `SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error)` — starts container, waits for readiness, runs migrations from `migrations/`
- [ ] 2.4 Implement `NewTestDB(container *PostgresContainer) *gorm.DB` — returns a shared `*gorm.DB` connected to the container
- [ ] 2.5 Implement `NewTestTx(t *testing.T, db *gorm.DB) *gorm.DB` — begins a transaction, registers rollback via `t.Cleanup`

## 3. Validation

- [ ] 3.1 Write a smoke test in `internal/testutil/` that calls `SetupPostgresContainer`, `NewTestDB`, and `NewTestTx` to confirm the full lifecycle works end-to-end
