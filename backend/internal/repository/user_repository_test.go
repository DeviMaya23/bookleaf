package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserRepository_GetOrCreate_HappyPath(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatalf("auto migrate users: %v", err)
	}

	repo := NewUserRepository(db)

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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	repo := NewUserRepository(db)

	_, err = repo.GetOrCreate(context.Background(), "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
