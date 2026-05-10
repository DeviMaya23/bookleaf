package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestFolderRepository_Create_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	repo := NewFolderRepository(tx)

	folder, err := repo.Create(context.Background(), &domain.Folder{
		UserID:      "kp_abc123",
		Name:        "travel",
		Description: func() *string { v := "trip board"; return &v }(),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder.ID == uuid.Nil {
		t.Fatal("expected generated folder ID")
	}
	if folder.Name != "travel" {
		t.Fatalf("expected folder name travel, got: %s", folder.Name)
	}
	if folder.Description == nil || *folder.Description != "trip board" {
		t.Fatalf("expected folder description trip board, got: %v", folder.Description)
	}
}

func TestFolderRepository_Create_Failure(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderRepository(tx)

	_, err := repo.Create(context.Background(), &domain.Folder{
		UserID: "kp_missing",
		Name:   "travel",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFolderRepository_List_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	createFolder(t, tx, "kp_abc123", "travel", nil)
	createFolder(t, tx, "kp_abc123", "food", nil)
	repo := NewFolderRepository(tx)

	folders, err := repo.List(context.Background(), "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(folders) != 2 {
		t.Fatalf("expected 2 folders, got: %d", len(folders))
	}
}

func TestFolderRepository_List_Failure(t *testing.T) {
	db, err := testutil.NewTestDB(testContainer)
	if err != nil {
		t.Fatalf("create db handle: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying sql.DB: %v", err)
	}
	sqlDB.Close()

	repo := NewFolderRepository(db)

	_, err = repo.List(context.Background(), "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFolderRepository_FindByName_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	existing := createFolder(t, tx, "kp_abc123", "Nature", nil)
	repo := NewFolderRepository(tx)

	folder, err := repo.FindByName(context.Background(), "kp_abc123", "nature")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder == nil {
		t.Fatal("expected folder, got nil")
	}
	if folder.ID != existing.ID {
		t.Fatalf("expected id %s, got %s", existing.ID, folder.ID)
	}
}

func TestFolderRepository_FindByName_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	repo := NewFolderRepository(tx)

	folder, err := repo.FindByName(context.Background(), "kp_abc123", "missing-folder")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder != nil {
		t.Fatalf("expected nil folder, got: %+v", folder)
	}
}

func TestFolderRepository_GetByID_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	existing := createFolder(t, tx, "kp_abc123", "travel", nil)
	repo := NewFolderRepository(tx)

	folder, err := repo.GetByID(context.Background(), existing.ID, "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder.ID != existing.ID {
		t.Fatalf("expected id %s, got %s", existing.ID, folder.ID)
	}
}

func TestFolderRepository_GetByID_Failure(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderRepository(tx)

	_, err := repo.GetByID(context.Background(), uuid.New(), "kp_abc123")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestFolderRepository_Update_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	parent := createFolder(t, tx, "kp_abc123", "parent", nil)
	existing := createFolder(t, tx, "kp_abc123", "travel", nil)
	repo := NewFolderRepository(tx)

	updated, err := repo.Update(context.Background(), &domain.Folder{
		ID:          existing.ID,
		UserID:      "kp_abc123",
		Name:        "updated",
		ParentID:    &parent.ID,
		Description: func() *string { v := "updated description"; return &v }(),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Name != "updated" {
		t.Fatalf("expected name updated, got %s", updated.Name)
	}
	if updated.ParentID == nil || *updated.ParentID != parent.ID {
		t.Fatalf("expected parent_id %s, got %v", parent.ID, updated.ParentID)
	}
	if updated.Description == nil || *updated.Description != "updated description" {
		t.Fatalf("expected description updated description, got %v", updated.Description)
	}
}

func TestFolderRepository_Update_Failure(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderRepository(tx)

	_, err := repo.Update(context.Background(), &domain.Folder{
		ID:     uuid.New(),
		UserID: "kp_abc123",
		Name:   "updated",
	})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestFolderRepository_DeleteWithCascade_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userID := "kp_abc123"
	createUser(t, tx, userID)
	target := createFolder(t, tx, userID, "target", nil)
	child := createFolder(t, tx, userID, "child", &target.ID)
	image := createImage(t, tx, userID, &target.ID)
	repo := NewFolderRepository(tx)

	err := repo.DeleteWithCascade(context.Background(), target.ID, userID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var gotTarget domain.Folder
	err = tx.Where("id = ?", target.ID).First(&gotTarget).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected target folder deleted, got error: %v", err)
	}

	var gotChild domain.Folder
	if err := tx.Where("id = ?", child.ID).First(&gotChild).Error; err != nil {
		t.Fatalf("expected child folder to exist, got: %v", err)
	}
	if gotChild.ParentID != nil {
		t.Fatalf("expected child parent_id nil, got: %v", gotChild.ParentID)
	}

	var gotImage domain.Image
	if err := tx.Where("id = ?", image.ID).First(&gotImage).Error; err != nil {
		t.Fatalf("expected image to exist, got: %v", err)
	}
	if gotImage.FolderID != nil {
		t.Fatalf("expected image folder_id nil, got: %v", gotImage.FolderID)
	}
}

func TestFolderRepository_DeleteWithCascade_FolderOwnedByAnotherUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	target := createFolder(t, tx, "kp_owner", "target", nil)
	repo := NewFolderRepository(tx)

	err := repo.DeleteWithCascade(context.Background(), target.ID, "kp_other")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}

	var got domain.Folder
	if err := tx.Where("id = ?", target.ID).First(&got).Error; err != nil {
		t.Fatalf("expected folder to remain for owner, got: %v", err)
	}
}

func createUser(t *testing.T, db *gorm.DB, id string) {
	t.Helper()
	if err := db.Create(&domain.User{ID: id}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func createFolder(t *testing.T, db *gorm.DB, userID, name string, parentID *uuid.UUID) *domain.Folder {
	t.Helper()
	folder := &domain.Folder{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	}
	if err := db.Create(folder).Error; err != nil {
		t.Fatalf("create folder: %v", err)
	}
	return folder
}

func createImage(t *testing.T, db *gorm.DB, userID string, folderID *uuid.UUID) *domain.Image {
	t.Helper()
	image := &domain.Image{
		UserID:   userID,
		FolderID: folderID,
		Title:    "test image",
		R2Path:   "users/kp_abc123/images/test.jpg",
		MIMEType: "image/jpeg",
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image
}
