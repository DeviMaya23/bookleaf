package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/testutil"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupImageTest(t *testing.T) (usecase.ImageRepository, string) {
	t.Helper()
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_imgtest")
	require.NoError(t, err)

	return NewImageRepository(tx), user.ID
}

func newTestImage(userID string) *domain.Image {
	id := uuid.New()
	return &domain.Image{
		ID:       id,
		UserID:   userID,
		Title:    "test image",
		MIMEType: "image/jpeg",
		R2Path:   "users/" + userID + "/images/" + id.String() + ".jpg",
	}
}

func TestImageRepository_Create_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	img, err := repo.Create(context.Background(), newTestImage(userID))

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, img.ID)
	assert.Equal(t, userID, img.UserID)
	assert.Equal(t, "test image", img.Title)
}

func TestImageRepository_Create_DBError(t *testing.T) {
	db, err := testutil.NewTestDB(testContainer)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	repo := NewImageRepository(db)
	_, err = repo.Create(context.Background(), newTestImage("kp_imgtest"))
	require.Error(t, err)
}

func TestImageRepository_List_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), userID, nil)

	require.NoError(t, err)
	assert.Len(t, images, 2)
}

func TestImageRepository_List_FilterByFolder(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_imgtest")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "moodboard"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	img1 := newTestImage(user.ID)
	img1.FolderID = &folder.ID
	_, err = repo.Create(context.Background(), img1)
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), user.ID, &folder.ID)

	require.NoError(t, err)
	assert.Len(t, images, 1)
	assert.Equal(t, &folder.ID, images[0].FolderID)
}

func TestImageRepository_GetByID_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestImageRepository_GetByID_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.GetByID(context.Background(), uuid.New(), userID)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_UpdateThumbnailPath_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	thumbPath := "users/" + userID + "/thumbnails/" + created.ID.String() + ".jpg"
	err = repo.UpdateThumbnailPath(context.Background(), created.ID, thumbPath)

	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, found.ThumbnailPath)
	assert.Equal(t, thumbPath, *found.ThumbnailPath)
}

func TestImageRepository_UpdateThumbnailPath_NotFound(t *testing.T) {
	repo, _ := setupImageTest(t)

	err := repo.UpdateThumbnailPath(context.Background(), uuid.New(), "some/path.jpg")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_SoftDelete_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_SoftDelete_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	err := repo.SoftDelete(context.Background(), uuid.New(), userID)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_GetDeletedByID_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	deleted, err := repo.GetDeletedByID(context.Background(), created.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, deleted.ID)
	assert.True(t, deleted.DeletedAt.Valid)
}

func TestImageRepository_GetDeletedByID_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	// non-deleted image should not be returned by GetDeletedByID
	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	_, err = repo.GetDeletedByID(context.Background(), created.ID, userID)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_Restore_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	err = repo.Restore(context.Background(), created.ID, userID)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.False(t, found.DeletedAt.Valid)
}

func TestImageRepository_Restore_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	err := repo.Restore(context.Background(), uuid.New(), userID)

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_ListTrashed_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	img1, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	img2, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), img1.ID, userID)
	require.NoError(t, err)
	err = repo.SoftDelete(context.Background(), img2.ID, userID)
	require.NoError(t, err)

	// create one non-deleted image to confirm it's excluded
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	trashed, err := repo.ListTrashed(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, trashed, 2)
}

func TestImageRepository_ListTrashed_Empty(t *testing.T) {
	repo, userID := setupImageTest(t)

	trashed, err := repo.ListTrashed(context.Background(), userID)

	require.NoError(t, err)
	assert.Empty(t, trashed)
}

func TestImageRepository_Update_SelectiveFieldUpdate(t *testing.T) {
	repo, userID := setupImageTest(t)

	img := newTestImage(userID)
	thumbPath := "users/" + userID + "/thumbnails/test.jpg"
	img.ThumbnailPath = &thumbPath
	created, err := repo.Create(context.Background(), img)
	require.NoError(t, err)

	updated, err := repo.Update(context.Background(), created.ID, userID, map[string]any{
		"title": "new title",
	})

	require.NoError(t, err)
	assert.Equal(t, "new title", updated.Title)
	require.NotNil(t, updated.ThumbnailPath)
	assert.Equal(t, thumbPath, *updated.ThumbnailPath)
}

func TestImageRepository_Update_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.Update(context.Background(), uuid.New(), userID, map[string]any{
		"title": "ghost",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
