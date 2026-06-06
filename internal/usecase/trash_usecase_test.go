package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- test doubles ---

type mockTrashJobEnqueuer struct {
	insertCalls int
	insertedArgs []JobArgs
	err         error
}

func (m *mockTrashJobEnqueuer) Insert(_ context.Context, args JobArgs) error {
	m.insertCalls++
	m.insertedArgs = append(m.insertedArgs, args)
	return m.err
}

// --- helpers ---

func newTrashUsecase(imageRepo ImageRepository, store StorageService, enqueuer JobEnqueuer) *trashUsecase {
	return NewTrashUsecase(imageRepo, store, enqueuer, noopTel())
}

// --- EmptyTrash ---

func TestTrashUsecase_EmptyTrash_Success(t *testing.T) {
	thumbnailPath := "users/kp_u1/thumbnails/b.jpg"
	images := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg", ThumbnailPath: &thumbnailPath},
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/b.jpg"},
	}
	repo := &mockImageRepository{images: images}
	store := &mockStorageService{}
	enqueuer := &mockTrashJobEnqueuer{}
	uc := newTrashUsecase(repo, store, enqueuer)

	err := uc.EmptyTrash(context.Background(), "kp_u1")

	require.NoError(t, err)
	assert.Equal(t, 2, repo.hardDeleteCalls)
	assert.Equal(t, 0, store.deleteCalls)
	assert.Equal(t, 2, enqueuer.insertCalls)
}

func TestTrashUsecase_EmptyTrash_NoOp(t *testing.T) {
	repo := &mockImageRepository{images: []*domain.Image{}}
	enqueuer := &mockTrashJobEnqueuer{}
	uc := newTrashUsecase(repo, &mockStorageService{}, enqueuer)

	err := uc.EmptyTrash(context.Background(), "kp_u1")

	require.NoError(t, err)
	assert.Equal(t, 0, repo.hardDeleteCalls)
	assert.Equal(t, 0, enqueuer.insertCalls)
}

// --- DeleteFromTrash ---

func TestTrashUsecase_DeleteFromTrash_Success(t *testing.T) {
	imageID := uuid.New()
	thumbnailPath := "users/kp_u1/thumbnails/b.jpg"
	img := &domain.Image{ID: imageID, UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg", ThumbnailPath: &thumbnailPath}
	repo := &mockImageRepository{image: img}
	store := &mockStorageService{}
	uc := newTrashUsecase(repo, store, &mockTrashJobEnqueuer{})

	err := uc.DeleteFromTrash(context.Background(), imageID, "kp_u1")

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, img.R2Path)
	assert.Contains(t, store.deletedKeys, thumbnailPath)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}

func TestTrashUsecase_DeleteFromTrash_NotInTrash(t *testing.T) {
	repo := &mockImageRepository{err: gorm.ErrRecordNotFound}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockTrashJobEnqueuer{})

	err := uc.DeleteFromTrash(context.Background(), uuid.New(), "kp_u1")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTrashUsecase_DeleteFromTrash_StorageErrorContinues(t *testing.T) {
	imageID := uuid.New()
	img := &domain.Image{ID: imageID, UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"}
	repo := &mockImageRepository{image: img}
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newTrashUsecase(repo, store, &mockTrashJobEnqueuer{})

	err := uc.DeleteFromTrash(context.Background(), imageID, "kp_u1")

	require.NoError(t, err)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}

// --- ProcessR2Delete ---

func TestTrashUsecase_ProcessR2Delete_WithThumbnail(t *testing.T) {
	thumbnailPath := "users/kp_u1/thumbnails/b.jpg"
	store := &mockStorageService{}
	uc := newTrashUsecase(&mockImageRepository{}, store, &mockTrashJobEnqueuer{})

	err := uc.ProcessR2Delete(context.Background(), "users/kp_u1/images/a.jpg", &thumbnailPath)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_u1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, thumbnailPath)
}

func TestTrashUsecase_ProcessR2Delete_WithoutThumbnail(t *testing.T) {
	store := &mockStorageService{}
	uc := newTrashUsecase(&mockImageRepository{}, store, &mockTrashJobEnqueuer{})

	err := uc.ProcessR2Delete(context.Background(), "users/kp_u1/images/a.jpg", nil)

	require.NoError(t, err)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_u1/images/a.jpg")
}

func TestTrashUsecase_ProcessR2Delete_StorageError(t *testing.T) {
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newTrashUsecase(&mockImageRepository{}, store, &mockTrashJobEnqueuer{})

	err := uc.ProcessR2Delete(context.Background(), "users/kp_u1/images/a.jpg", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete r2 object")
}

// --- PurgeExpiredTrash ---

func TestTrashUsecase_PurgeExpiredTrash_WithThumbnail(t *testing.T) {
	thumbnailPath := "users/kp_u1/thumbnails/b.jpg"
	expired := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg", ThumbnailPath: &thumbnailPath},
	}
	repo := &mockImageRepository{images: expired}
	store := &mockStorageService{}
	uc := newTrashUsecase(repo, store, &mockTrashJobEnqueuer{})

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_u1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, thumbnailPath)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}

func TestTrashUsecase_PurgeExpiredTrash_StorageErrorContinues(t *testing.T) {
	expired := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
	}
	repo := &mockImageRepository{images: expired}
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newTrashUsecase(repo, store, &mockTrashJobEnqueuer{})

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.NoError(t, err)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}
