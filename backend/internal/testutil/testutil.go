package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresContainer wraps a running testcontainers Postgres instance.
type PostgresContainer struct {
	inner   *tcpostgres.PostgresContainer
	ConnStr string
}

// Terminate stops and removes the container. Call this in TestMain after m.Run().
func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.inner.Terminate(ctx)
}

// SetupPostgresContainer starts a Postgres container and runs all migrations
// from the nearest migration/ folder (found by walking up from the working directory).
func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	inner, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategyAndDeadline(30*time.Second,
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connStr, err := inner.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = inner.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	migrationDir, err := findMigrationDir()
	if err != nil {
		_ = inner.Terminate(ctx)
		return nil, fmt.Errorf("find migration directory: %w", err)
	}

	m, err := migrate.New("file://"+migrationDir, connStr)
	if err != nil {
		_ = inner.Terminate(ctx)
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		_ = inner.Terminate(ctx)
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &PostgresContainer{inner: inner, ConnStr: connStr}, nil
}

// NewTestDB returns a *gorm.DB connected to the container. Safe to share across tests.
func NewTestDB(container *PostgresContainer) (*gorm.DB, error) {
	db, err := gorm.Open(gormpostgres.Open(container.ConnStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open gorm connection: %w", err)
	}
	return db, nil
}

// NewTestTx returns a transaction-scoped *gorm.DB. The transaction is automatically
// rolled back when the test finishes via t.Cleanup, keeping state isolated between tests.
func NewTestTx(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()
	tx := db.Begin()
	t.Cleanup(func() {
		tx.Rollback()
	})
	return tx
}

// findMigrationDir walks up from the current working directory until it finds
// a go.mod file, then returns the sibling migration/ directory.
func findMigrationDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			migDir := filepath.Join(dir, "migration")
			if _, err := os.Stat(migDir); err != nil {
				return "", fmt.Errorf("migration directory not found at %s", migDir)
			}
			return migDir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found; cannot locate migration directory")
		}
		dir = parent
	}
}
