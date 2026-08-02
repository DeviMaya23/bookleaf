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
	image              *domain.Image
	images             []*domain.Image
	err                error
	hardDeleteCalls    int
	lastListName       *string
	lastListSort       *string
	lastListDirection  *string
	filterOwnedCalls   int
	lastFilterOwnedIDs []uuid.UUID
	filterOwnedResult  []uuid.UUID
	softDeleteFailIDs  map[uuid.UUID]struct{}
	softDeleteCalls    int
	lastSoftDeleteIDs  []uuid.UUID
}

func (m *mockTrashRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}
func (m *mockTrashRepository) GetDeletedByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}
func (m *mockTrashRepository) SoftDelete(_ context.Context, id uuid.UUID, _ string) error {
	m.softDeleteCalls++
	m.lastSoftDeleteIDs = append(m.lastSoftDeleteIDs, id)
	if _, fail := m.softDeleteFailIDs[id]; fail {
		return gorm.ErrRecordNotFound
	}
	return m.err
}
func (m *mockTrashRepository) Restore(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockTrashRepository) ListTrashed(_ context.Context, _ string, name *string, sortField *string, direction *string, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	m.lastListName = name
	m.lastListSort = sortField
	m.lastListDirection = direction
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
func (m *mockTrashRepository) FilterOwnedImageIDs(_ context.Context, ids []uuid.UUID, _ string) ([]uuid.UUID, error) {
	m.filterOwnedCalls++
	m.lastFilterOwnedIDs = ids
	if m.filterOwnedResult != nil {
		return m.filterOwnedResult, m.err
	}
	return ids, m.err
}

type mockJobEnqueuer struct {
	err              error
	insertArgs       []JobArgs
	insertUniqueArgs []JobArgs
}

func (m *mockJobEnqueuer) Insert(_ context.Context, args JobArgs) error {
	m.insertArgs = append(m.insertArgs, args)
	return m.err
}

func (m *mockJobEnqueuer) InsertUnique(_ context.Context, args JobArgs) error {
	m.insertUniqueArgs = append(m.insertUniqueArgs, args)
	return m.err
}

func newTrashUsecase(repo TrashRepository, store StorageService, enqueuer JobEnqueuer) *trashUsecase {
	return NewTrashUsecase(repo, store, enqueuer, noopTel())
}

// --- ListTrashed ---

func TestTrashUsecase_ListTrashed_PassesNameToRepository(t *testing.T) {
	repo := &mockTrashRepository{images: []*domain.Image{{ID: uuid.New()}}}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})
	name := "heartopia"

	_, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{Name: &name})

	require.NoError(t, err)
	require.NotNil(t, repo.lastListName)
	assert.Equal(t, "heartopia", *repo.lastListName)
}

func TestTrashUsecase_ListTrashed_SkipsBlankName(t *testing.T) {
	empty := ""
	tests := []struct {
		name   string
		params ListTrashedParams
	}{
		{name: "nil name", params: ListTrashedParams{}},
		{name: "empty string name", params: ListTrashedParams{Name: &empty}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrashRepository{images: []*domain.Image{{ID: uuid.New()}}}
			uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

			_, err := uc.ListTrashed(context.Background(), "kp_abc123", tt.params)

			require.NoError(t, err)
			assert.Nil(t, repo.lastListName)
		})
	}
}

func TestTrashUsecase_ListTrashed_PassesSortAndDirection(t *testing.T) {
	sortVal := "title"
	dirVal := "asc"

	repo := &mockTrashRepository{images: []*domain.Image{}}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	_, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{
		Sort:      &sortVal,
		Direction: &dirVal,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.lastListSort)
	assert.Equal(t, "title", *repo.lastListSort)
	require.NotNil(t, repo.lastListDirection)
	assert.Equal(t, "asc", *repo.lastListDirection)
}

func TestTrashUsecase_ListTrashed_PassesNilSortAndDirection(t *testing.T) {
	repo := &mockTrashRepository{images: []*domain.Image{}}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	_, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{
		Sort:      nil,
		Direction: nil,
	})

	require.NoError(t, err)
	assert.Nil(t, repo.lastListSort)
	assert.Nil(t, repo.lastListDirection)
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

// --- BulkTrash ---

func TestTrashUsecase_BulkTrash_AllSucceed(t *testing.T) {
	imageIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	repo := &mockTrashRepository{filterOwnedResult: imageIDs}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	count, err := uc.BulkTrash(context.Background(), "u1", imageIDs)

	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, repo.softDeleteCalls)
}

func TestTrashUsecase_BulkTrash_AlreadyTrashedExcludedOthersSucceed(t *testing.T) {
	alreadyTrashed := uuid.New()
	valid := uuid.New()
	repo := &mockTrashRepository{
		filterOwnedResult: []uuid.UUID{alreadyTrashed, valid},
		softDeleteFailIDs: map[uuid.UUID]struct{}{alreadyTrashed: {}},
	}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	count, err := uc.BulkTrash(context.Background(), "u1", []uuid.UUID{alreadyTrashed, valid})

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 2, repo.softDeleteCalls)
}

func TestTrashUsecase_BulkTrash_UnownedImageExcludedOthersSucceed(t *testing.T) {
	owned := uuid.New()
	unowned := uuid.New()
	repo := &mockTrashRepository{filterOwnedResult: []uuid.UUID{owned}}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	count, err := uc.BulkTrash(context.Background(), "u1", []uuid.UUID{owned, unowned})

	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, repo.softDeleteCalls)
	assert.Equal(t, []uuid.UUID{owned}, repo.lastSoftDeleteIDs)
}

func TestTrashUsecase_BulkTrash_AllInvalidReturnsZeroNoError(t *testing.T) {
	imageIDs := []uuid.UUID{uuid.New(), uuid.New()}
	repo := &mockTrashRepository{filterOwnedResult: []uuid.UUID{}}
	uc := newTrashUsecase(repo, &mockStorageService{}, &mockJobEnqueuer{})

	count, err := uc.BulkTrash(context.Background(), "u1", imageIDs)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, repo.softDeleteCalls)
}
