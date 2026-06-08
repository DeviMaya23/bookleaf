package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- Create ---

func TestFolderRepository_Create_PersistsFields(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	repo := NewFolderRepository(tx)

	folder, err := repo.Create(context.Background(), &domain.Folder{
		UserID:      "kp_abc123",
		Name:        "travel",
		Description: func() *string { v := "trip board"; return &v }(),
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, folder.ID)
	assert.Equal(t, "travel", folder.Name)
	require.NotNil(t, folder.Description)
	assert.Equal(t, "trip board", *folder.Description)
}

func TestFolderRepository_Create_FKViolation(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderRepository(tx)

	_, err := repo.Create(context.Background(), &domain.Folder{
		UserID: "kp_nonexistent",
		Name:   "travel",
	})

	require.Error(t, err)
}

// --- List ---

func TestFolderRepository_List_ReturnsOwnFolders(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	createFolder(t, tx, "kp_abc123", "travel", nil)
	createFolder(t, tx, "kp_abc123", "food", nil)
	repo := NewFolderRepository(tx)

	folders, err := repo.List(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.Len(t, folders, 2)
}

func TestFolderRepository_List_ExcludesOtherUserFolders(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	createFolder(t, tx, "kp_owner", "travel", nil)
	repo := NewFolderRepository(tx)

	folders, err := repo.List(context.Background(), "kp_other")

	require.NoError(t, err)
	assert.Empty(t, folders)
}

// --- FindByName ---

func TestFolderRepository_FindByName_ReturnsMatch(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	existing := createFolder(t, tx, "kp_abc123", "Nature", nil)
	repo := NewFolderRepository(tx)

	folder, err := repo.FindByName(context.Background(), "kp_abc123", "nature")

	require.NoError(t, err)
	require.NotNil(t, folder)
	assert.Equal(t, existing.ID, folder.ID)
}

func TestFolderRepository_FindByName_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	repo := NewFolderRepository(tx)

	folder, err := repo.FindByName(context.Background(), "kp_abc123", "missing-folder")

	require.NoError(t, err)
	assert.Nil(t, folder)
}

func TestFolderRepository_FindByName_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	createFolder(t, tx, "kp_owner", "nature", nil)
	repo := NewFolderRepository(tx)

	folder, err := repo.FindByName(context.Background(), "kp_other", "nature")

	require.NoError(t, err)
	assert.Nil(t, folder)
}

// --- GetByID ---

func TestFolderRepository_GetByID_ReturnsFolder(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	existing := createFolder(t, tx, "kp_abc123", "travel", nil)
	repo := NewFolderRepository(tx)

	folder, err := repo.GetByID(context.Background(), existing.ID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, existing.ID, folder.ID)
}

func TestFolderRepository_GetByID_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderRepository(tx)

	_, err := repo.GetByID(context.Background(), uuid.New(), "kp_abc123")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestFolderRepository_GetByID_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	existing := createFolder(t, tx, "kp_owner", "travel", nil)
	repo := NewFolderRepository(tx)

	_, err := repo.GetByID(context.Background(), existing.ID, "kp_other")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// --- Update ---

func setupUpdateFolder(t *testing.T, tx *gorm.DB, userID string, parentID *uuid.UUID, description *string) *domain.Folder {
	t.Helper()
	folder := &domain.Folder{
		UserID:      userID,
		Name:        "travel",
		ParentID:    parentID,
		Description: description,
	}
	require.NoError(t, tx.Create(folder).Error)
	return folder
}

func TestFolderRepository_Update_OmittedFieldsPreserved(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	parent := createFolder(t, tx, "kp_abc123", "parent", nil)
	description := "original description"
	existing := setupUpdateFolder(t, tx, "kp_abc123", &parent.ID, &description)
	repo := NewFolderRepository(tx)

	updated, err := repo.Update(context.Background(), existing.ID, "kp_abc123", map[string]any{
		"name": "updated",
	})

	require.NoError(t, err)
	assert.Equal(t, "updated", updated.Name)
	require.NotNil(t, updated.ParentID)
	assert.Equal(t, parent.ID, *updated.ParentID)
	require.NotNil(t, updated.Description)
	assert.Equal(t, "original description", *updated.Description)
}

func TestFolderRepository_Update_ExplicitNullClearsField(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	parent := createFolder(t, tx, "kp_abc123", "parent", nil)
	description := "original description"
	existing := setupUpdateFolder(t, tx, "kp_abc123", &parent.ID, &description)
	repo := NewFolderRepository(tx)

	updated, err := repo.Update(context.Background(), existing.ID, "kp_abc123", map[string]any{
		"parent_id":   (*uuid.UUID)(nil),
		"description": (*string)(nil),
	})

	require.NoError(t, err)
	assert.Equal(t, "travel", updated.Name)
	assert.Nil(t, updated.ParentID)
	assert.Nil(t, updated.Description)
}

func TestFolderRepository_Update_OverwritesProvidedFields(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	oldParent := createFolder(t, tx, "kp_abc123", "old-parent", nil)
	newParent := createFolder(t, tx, "kp_abc123", "new-parent", nil)
	description := "original description"
	existing := setupUpdateFolder(t, tx, "kp_abc123", &oldParent.ID, &description)
	repo := NewFolderRepository(tx)

	updated, err := repo.Update(context.Background(), existing.ID, "kp_abc123", map[string]any{
		"name":        "updated",
		"parent_id":   &newParent.ID,
		"description": func() *string { v := "updated description"; return &v }(),
	})

	require.NoError(t, err)
	assert.Equal(t, "updated", updated.Name)
	require.NotNil(t, updated.ParentID)
	assert.Equal(t, newParent.ID, *updated.ParentID)
	require.NotNil(t, updated.Description)
	assert.Equal(t, "updated description", *updated.Description)
}

func TestFolderRepository_Update_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderRepository(tx)

	_, err := repo.Update(context.Background(), uuid.New(), "kp_abc123", map[string]any{
		"name": "updated",
	})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestFolderRepository_Update_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	existing := createFolder(t, tx, "kp_owner", "travel", nil)
	repo := NewFolderRepository(tx)

	_, err := repo.Update(context.Background(), existing.ID, "kp_other", map[string]any{
		"name": "updated",
	})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// --- DeleteWithCascade ---

func TestFolderRepository_DeleteWithCascade_RemovesTargetAndNullsChildParent(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userID := "kp_abc123"
	createUser(t, tx, userID)
	target := createFolder(t, tx, userID, "target", nil)
	child := createFolder(t, tx, userID, "child", &target.ID)
	img := createImage(t, tx, userID, &target.ID)
	repo := NewFolderRepository(tx)

	err := repo.DeleteWithCascade(context.Background(), target.ID, userID)

	require.NoError(t, err)

	var gotTarget domain.Folder
	require.ErrorIs(t, tx.Where("id = ?", target.ID).First(&gotTarget).Error, gorm.ErrRecordNotFound)

	var gotChild domain.Folder
	require.NoError(t, tx.Where("id = ?", child.ID).First(&gotChild).Error)
	assert.Nil(t, gotChild.ParentID)

	// image_folders row removed via ON DELETE CASCADE on folder_id FK
	var count int64
	tx.Table("image_folders").Where("image_id = ?", img.ID).Count(&count)
	assert.EqualValues(t, 0, count)

	// image itself survives — only the folder membership is removed
	var gotImage domain.Image
	require.NoError(t, tx.Where("id = ?", img.ID).First(&gotImage).Error)
}

func TestFolderRepository_DeleteWithCascade_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	target := createFolder(t, tx, "kp_owner", "target", nil)
	repo := NewFolderRepository(tx)

	err := repo.DeleteWithCascade(context.Background(), target.ID, "kp_other")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var got domain.Folder
	require.NoError(t, tx.Where("id = ?", target.ID).First(&got).Error)
}

// --- CountImagesByFolder ---

func TestFolderRepository_CountImagesByFolder_ExcludesSoftDeleted(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userID := "kp_countimgs"
	createUser(t, tx, userID)
	folder := createFolder(t, tx, userID, "target", nil)
	imageRepo := NewImageRepository(tx)
	createImage(t, tx, userID, &folder.ID)
	img2 := createImage(t, tx, userID, &folder.ID)
	createImage(t, tx, userID, nil) // unfiled — should not count
	require.NoError(t, imageRepo.SoftDelete(context.Background(), img2.ID, userID))
	repo := NewFolderRepository(tx)

	count, err := repo.CountImagesByFolder(context.Background(), folder.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestFolderRepository_CountImagesByFolder_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_owner")
	createUser(t, tx, "kp_other")
	folder := createFolder(t, tx, "kp_owner", "target", nil)
	createImage(t, tx, "kp_owner", &folder.ID)
	repo := NewFolderRepository(tx)

	count, err := repo.CountImagesByFolder(context.Background(), folder.ID, "kp_other")

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// --- helpers ---

func createUser(t *testing.T, db *gorm.DB, id string) {
	t.Helper()
	require.NoError(t, db.Create(&domain.User{ID: id}).Error)
}

func createFolder(t *testing.T, db *gorm.DB, userID, name string, parentID *uuid.UUID) *domain.Folder {
	t.Helper()
	folder := &domain.Folder{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	}
	require.NoError(t, db.Create(folder).Error)
	return folder
}

func createImage(t *testing.T, db *gorm.DB, userID string, folderID *uuid.UUID) *domain.Image {
	t.Helper()
	img := &domain.Image{
		UserID:   userID,
		Title:    "test image",
		R2Path:   "users/" + userID + "/images/test.jpg",
		MIMEType: "image/jpeg",
	}
	require.NoError(t, db.Create(img).Error)
	if folderID != nil {
		require.NoError(t, NewImageRepository(db).SetImageFolder(context.Background(), img.ID, folderID))
	}
	return img
}

