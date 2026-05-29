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

func setupTagTest(t *testing.T) (usecase.TagRepository, usecase.ImageRepository, string) {
	t.Helper()
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_tagtest")
	require.NoError(t, err)

	return NewTagRepository(tx), NewImageRepository(tx), user.ID
}

func newTestTag(userID, name string) *domain.Tag {
	return &domain.Tag{
		UserID: userID,
		Name:   name,
	}
}

func TestTagRepository_Create_Success(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	tag, err := repo.Create(context.Background(), newTestTag(userID, "landscape"))

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tag.ID)
	assert.Equal(t, "landscape", tag.Name)
	assert.Equal(t, userID, tag.UserID)
}

func TestTagRepository_Create_DuplicateName(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	_, err := repo.Create(context.Background(), newTestTag(userID, "nature"))
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), newTestTag(userID, "nature"))
	require.Error(t, err)
}

func TestTagRepository_ListByUserID_Success(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	_, err := repo.Create(context.Background(), newTestTag(userID, "travel"))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestTag(userID, "art"))
	require.NoError(t, err)

	tags, err := repo.ListByUserID(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, tags, 2)
	assert.Equal(t, "art", tags[0].Name)
	assert.Equal(t, "travel", tags[1].Name)
}

func TestTagRepository_GetByID_Success(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	created, err := repo.Create(context.Background(), newTestTag(userID, "portrait"))
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "portrait", found.Name)
}

func TestTagRepository_GetByID_NotFound(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	_, err := repo.GetByID(context.Background(), uuid.New(), userID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTagRepository_Update_Success(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	created, err := repo.Create(context.Background(), newTestTag(userID, "oldname"))
	require.NoError(t, err)

	updated, err := repo.Update(context.Background(), created.ID, userID, "newname")

	require.NoError(t, err)
	assert.Equal(t, "newname", updated.Name)
}

func TestTagRepository_Update_NotFound(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	_, err := repo.Update(context.Background(), uuid.New(), userID, "newname")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTagRepository_Delete_Success(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	created, err := repo.Create(context.Background(), newTestTag(userID, "todelete"))
	require.NoError(t, err)

	err = repo.Delete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, userID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTagRepository_Delete_NotFound(t *testing.T) {
	repo, _, userID := setupTagTest(t)

	err := repo.Delete(context.Background(), uuid.New(), userID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTagRepository_ReplaceImageTags_SetAndReplace(t *testing.T) {
	tagRepo, imageRepo, userID := setupTagTest(t)

	img, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	tag1, err := tagRepo.Create(context.Background(), newTestTag(userID, "tag1"))
	require.NoError(t, err)
	tag2, err := tagRepo.Create(context.Background(), newTestTag(userID, "tag2"))
	require.NoError(t, err)

	err = tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag1.ID, tag2.ID})
	require.NoError(t, err)

	// Replace with only tag1
	err = tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag1.ID})
	require.NoError(t, err)

	fetched, err := imageRepo.GetByID(context.Background(), img.ID, userID)
	require.NoError(t, err)
	require.Len(t, fetched.Tags, 1)
	assert.Equal(t, tag1.ID, fetched.Tags[0].ID)
}

func TestTagRepository_ReplaceImageTags_ClearAll(t *testing.T) {
	tagRepo, imageRepo, userID := setupTagTest(t)

	img, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	tag, err := tagRepo.Create(context.Background(), newTestTag(userID, "cleartag"))
	require.NoError(t, err)

	err = tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag.ID})
	require.NoError(t, err)

	err = tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{})
	require.NoError(t, err)

	fetched, err := imageRepo.GetByID(context.Background(), img.ID, userID)
	require.NoError(t, err)
	assert.Empty(t, fetched.Tags)
}
