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

type mockTrashRepository struct {
	image           *domain.Image
	images          []*domain.Image
	err             error
	hardDeleteCalls int
}

func (m *mockTrashRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}
func (m *mockTrashRepository) GetDeletedByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}
func (m *mockTrashRepository) SoftDelete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockTrashRepository) Restore(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockTrashRepository) ListTrashed(_ context.Context, _ string, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	return m.images, m.err
}
func (m *mockTrashRepository) ListAllTrashed(_ context.Context, _ string) ([]*domain.Image, error) {
	return m.images, m.err
}
func (m *mockTrashRepository) ListExpiredTrash(_ context.Context, _ time.Time) ([]*domain.Image, error) {
	return m.images, m.err
}
func (m *mockTrashRepository) HardDelete(_ context.Context, _ uuid.UUID, _ string) error {
	m.hardDeleteCalls++
	return m.err
}

type mockJobEnqueuer struct {
	err        error
	insertArgs []JobArgs
}

func (m *mockJobEnqueuer) Insert(_ context.Context, args JobArgs) error {
	m.insertArgs = append(m.insertArgs, args)
	return m.err
}

func newTrashUsecase(repo TrashRepository, store StorageService, enqueuer JobEnqueuer) *trashUsecase {
	return NewTrashUsecase(repo, store, enqueuer, noopTel())
}

// --- EmptyTrash ---

func TestTrashUsecase_EmptyTrash_DeletesAndEnqueues(t *testing.T) {
	r2Path := "users/u1/images/a.jpg"
	thumb := "users/u1/thumbnails/a.jpg"
	images := []*domain.Image{
		{ID: uuid.New(), UserID: "u1", R2Path: r2Path, ThumbnailPath: &thumb},
		{ID: uuid.New(), UserID: "u1", R2Path: "users/u1/images/b.jpg"},
	}
	repo := &mockTrashRepository{images: images}
	enqueuer := &mockJobEnqueuer{}
	uc := newTrashUsecase(repo, &mockStorageService{}, enqueuer)

	err := uc.EmptyTrash(context.Background(), "u1")

	require.NoError(t, err)
	assert.Equal(t, 2, repo.hardDeleteCalls)
	require.Len(t, enqueuer.insertArgs, 2)
	first := enqueuer.insertArgs[0].(R2DeleteArgs)
	assert.Equal(t, r2Path, first.R2Path)
	assert.Equal(t, thumb, *first.ThumbnailPath)
}

func TestTrashUsecase_EmptyTrash_NoOp(t *testing.T) {
	repo := &mockTrashRepository{images: nil}
	enqueuer := &mockJobEnqueuer{}
	uc := newTrashUsecase(repo, &mockStorageService{}, enqueuer)

	err := uc.EmptyTrash(context.Background(), "u1")

	require.NoError(t, err)
	assert.Equal(t, 0, repo.hardDeleteCalls)
	assert.Empty(t, enqueuer.insertArgs)
}

func TestTrashUsecase_EmptyTrash_EnqueueErrorContinues(t *testing.T) {
	images := []*domain.Image{
		{ID: uuid.New(), UserID: "u1", R2Path: "users/u1/images/a.jpg"},
		{ID: uuid.New(), UserID: "u1", R2Path: "users/u1/images/b.jpg"},
	}
	repo := &mockTrashRepository{images: images}
	enqueuer := &mockJobEnqueuer{err: errors.New("queue unavailable")}
	uc := newTrashUsecase(repo, &mockStorageService{}, enqueuer)

	err := uc.EmptyTrash(context.Background(), "u1")

	require.NoError(t, err)
	assert.Equal(t, 2, repo.hardDeleteCalls)
	// both enqueue attempts were made despite the first failing
	assert.Len(t, enqueuer.insertArgs, 2)
}

// --- DeleteFromTrash ---

func TestTrashUsecase_DeleteFromTrash_DeletesR2AndDB(t *testing.T) {
	r2Path := "users/u1/images/a.jpg"
	thumb := "users/u1/thumbnails/a.jpg"
	img := &domain.Image{ID: uuid.New(), UserID: "u1", R2Path: r2Path, ThumbnailPath: &thumb}
	repo := &mockTrashRepository{image: img}
	store := &mockStorageService{}
	uc := newTrashUsecase(repo, store, &mockJobEnqueuer{})

	err := uc.DeleteFromTrash(context.Background(), img.ID, "u1")

	require.NoError(t, err)
	assert.Equal(t, 1, repo.hardDeleteCalls)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, r2Path)
	assert.Contains(t, store.deletedKeys, thumb)
}

func TestTrashUsecase_DeleteFromTrash_NotFound(t *testing.T) {
	repo := &mockTrashRepository{err: gorm.ErrRecordNotFound}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	err := uc.DeleteFromTrash(context.Background(), uuid.New(), "u1")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Equal(t, 0, repo.hardDeleteCalls)
}

func TestTrashUsecase_DeleteFromTrash_R2FailureDoesNotBlockHardDelete(t *testing.T) {
	img := &domain.Image{ID: uuid.New(), UserID: "u1", R2Path: "users/u1/images/a.jpg"}
	repo := &mockTrashRepository{image: img}
	store := &mockStorageService{deleteObjectErr: errors.New("r2 error")}
	uc := newTrashUsecase(repo, store, &mockJobEnqueuer{})

	err := uc.DeleteFromTrash(context.Background(), img.ID, "u1")

	require.NoError(t, err)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}

// --- ProcessR2Delete ---

func TestTrashUsecase_ProcessR2Delete_WithThumbnail(t *testing.T) {
	r2Path := "users/u1/images/a.jpg"
	thumb := "users/u1/thumbnails/a.jpg"
	store := &mockStorageService{}
	uc := newTrashUsecase(&mockTrashRepository{}, store, &mockJobEnqueuer{})

	err := uc.ProcessR2Delete(context.Background(), r2Path, &thumb)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, r2Path)
	assert.Contains(t, store.deletedKeys, thumb)
}

func TestTrashUsecase_ProcessR2Delete_WithoutThumbnail(t *testing.T) {
	r2Path := "users/u1/images/a.jpg"
	store := &mockStorageService{}
	uc := newTrashUsecase(&mockTrashRepository{}, store, &mockJobEnqueuer{})

	err := uc.ProcessR2Delete(context.Background(), r2Path, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, []string{r2Path}, store.deletedKeys)
}

func TestTrashUsecase_ProcessR2Delete_StorageError(t *testing.T) {
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newTrashUsecase(&mockTrashRepository{}, store, &mockJobEnqueuer{})

	err := uc.ProcessR2Delete(context.Background(), "users/u1/images/a.jpg", nil)

	require.Error(t, err)
}
