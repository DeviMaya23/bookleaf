package testutil

import (
	"context"
	"testing"
)

func TestSetupPostgresContainer(t *testing.T) {
	ctx := context.Background()

	container, err := SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("SetupPostgresContainer: %v", err)
	}
	defer container.Terminate(ctx)

	db, err := NewTestDB(container)
	if err != nil {
		t.Fatalf("NewTestDB: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	tx := NewTestTx(t, db)
	if tx == nil {
		t.Fatal("NewTestTx returned nil")
	}
	if tx.Error != nil {
		t.Fatalf("NewTestTx: %v", tx.Error)
	}
}
