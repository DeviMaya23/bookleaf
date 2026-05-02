package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/testutil"
)

func TestUserRepository_GetOrCreate_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	user, err := repo.GetOrCreate(context.Background(), "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID != "kp_abc123" {
		t.Fatalf("expected user ID kp_abc123, got: %s", user.ID)
	}
	if user.VisionEnabled {
		t.Fatalf("expected vision_enabled false, got true")
	}
}

func TestUserRepository_GetOrCreate_DBError(t *testing.T) {
	db, err := testutil.NewTestDB(testContainer)
	if err != nil {
		t.Fatalf("create db handle: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying sql.DB: %v", err)
	}
	sqlDB.Close()

	repo := NewUserRepository(db)

	_, err = repo.GetOrCreate(context.Background(), "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
