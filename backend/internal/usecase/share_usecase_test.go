package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- spies ---

// stubShareFolderRepo is a value-return spy for ShareFolderRepository.
type stubShareFolderRepo struct {
	folder *domain.Folder
	err    error
}

func (s *stubShareFolderRepo) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Folder, error) {
	return s.folder, s.err
}

// stubShareImageRepo is a value-return spy for ShareImageRepository.
type stubShareImageRepo struct {
	images []*domain.Image
	err    error
}

func (s *stubShareImageRepo) ListByFolder(_ context.Context, _ string, _ uuid.UUID, _ *string, _ *string) ([]*domain.Image, error) {
	return s.images, s.err
}

// spyShareStorage is a value-return spy for StorageService.GeneratePresignedGetURL.
type spyShareStorage struct {
	StorageService
	url string
	err error
}

func (s *spyShareStorage) GeneratePresignedGetURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.url + "?key=" + key, nil
}

func newTestShareUsecase(folderShareRepo FolderShareRepository, folderRepo ShareFolderRepository, imageRepo ShareImageRepository, store StorageService) *shareUsecase {
	return NewShareUsecase(folderShareRepo, folderRepo, imageRepo, store, noopTel())
}

// --- CreateShare ---

func TestShareUsecase_CreateShare_NewShareCreated(t *testing.T) {
	folderID := uuid.New()
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	token, created, err := uc.CreateShare(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
	assert.True(t, created)
	assert.NotEmpty(t, token)
	assert.Equal(t, 1, folderShareRepo.createCalls)
}

func TestShareUsecase_CreateShare_IdempotentRepeat(t *testing.T) {
	folderID := uuid.New()
	existing := &domain.FolderShare{ID: uuid.New(), FolderID: folderID, Token: "existing-token"}
	folderShareRepo := newFakeFolderShareRepo(existing)
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	token, created, err := uc.CreateShare(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "existing-token", token)
	assert.Equal(t, 0, folderShareRepo.createCalls)
}

func TestShareUsecase_CreateShare_ConcurrentCreateFallback(t *testing.T) {
	folderID := uuid.New()
	winning := &domain.FolderShare{ID: uuid.New(), FolderID: folderID, Token: "winning-token"}
	folderShareRepo := newFakeFolderShareRepo(winning)
	folderShareRepo.firstGetByFolderIDNotFound = true
	folderShareRepo.failCreateWithUniqueViolation = true
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	token, created, err := uc.CreateShare(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "winning-token", token)
}

func TestShareUsecase_CreateShare_NotOwned(t *testing.T) {
	folderID := uuid.New()
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{err: gorm.ErrRecordNotFound}, &stubShareImageRepo{}, &spyShareStorage{})

	_, _, err := uc.CreateShare(context.Background(), folderID, "kp_abc123")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Equal(t, 0, folderShareRepo.createCalls)
}

// --- GetShare ---

func TestShareUsecase_GetShare_Found(t *testing.T) {
	folderID := uuid.New()
	existing := &domain.FolderShare{ID: uuid.New(), FolderID: folderID, Token: "existing-token"}
	folderShareRepo := newFakeFolderShareRepo(existing)
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	token, err := uc.GetShare(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "existing-token", token)
}

func TestShareUsecase_GetShare_NotShared(t *testing.T) {
	folderID := uuid.New()
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	_, err := uc.GetShare(context.Background(), folderID, "kp_abc123")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestShareUsecase_GetShare_NotOwned(t *testing.T) {
	folderID := uuid.New()
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{err: gorm.ErrRecordNotFound}, &stubShareImageRepo{}, &spyShareStorage{})

	_, err := uc.GetShare(context.Background(), folderID, "kp_abc123")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// --- DeleteShare ---

func TestShareUsecase_DeleteShare_RevokesExisting(t *testing.T) {
	folderID := uuid.New()
	existing := &domain.FolderShare{ID: uuid.New(), FolderID: folderID, Token: "existing-token"}
	folderShareRepo := newFakeFolderShareRepo(existing)
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	err := uc.DeleteShare(context.Background(), folderID, "kp_abc123")
	require.NoError(t, err)

	_, getErr := folderShareRepo.GetByFolderID(context.Background(), folderID)
	require.ErrorIs(t, getErr, gorm.ErrRecordNotFound)
}

func TestShareUsecase_DeleteShare_NoopWhenAbsent(t *testing.T) {
	folderID := uuid.New()
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{folder: &domain.Folder{ID: folderID}}, &stubShareImageRepo{}, &spyShareStorage{})

	err := uc.DeleteShare(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
}

func TestShareUsecase_DeleteShare_NotOwned(t *testing.T) {
	folderID := uuid.New()
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{err: gorm.ErrRecordNotFound}, &stubShareImageRepo{}, &spyShareStorage{})

	err := uc.DeleteShare(context.Background(), folderID, "kp_abc123")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// --- GetSharedFolder ---

func TestShareUsecase_GetSharedFolder_AssemblesFolderAndImages(t *testing.T) {
	folderID := uuid.New()
	notes := "trip board"
	thumbnailPath := "users/kp_abc123/thumbs/a.jpg"
	share := &domain.FolderShare{
		ID:       uuid.New(),
		FolderID: folderID,
		Token:    "tok_abc123",
		Folder: domain.Folder{
			ID:          folderID,
			UserID:      "kp_abc123",
			Name:        "travel",
			Description: &notes,
		},
	}
	folderShareRepo := newFakeFolderShareRepo(share)
	images := []*domain.Image{
		{ID: uuid.New(), Title: "first", R2Path: "users/kp_abc123/images/a.jpg", ThumbnailPath: &thumbnailPath},
		{ID: uuid.New(), Title: "second", R2Path: "users/kp_abc123/images/b.jpg"},
	}
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{}, &stubShareImageRepo{images: images}, &spyShareStorage{url: "https://r2.example.com/presigned"})

	result, err := uc.GetSharedFolder(context.Background(), "tok_abc123")

	require.NoError(t, err)
	assert.Equal(t, "travel", result.Name)
	require.NotNil(t, result.Notes)
	assert.Equal(t, "trip board", *result.Notes)
	require.Len(t, result.Images, 2)

	assert.Equal(t, "first", result.Images[0].Title)
	require.NotNil(t, result.Images[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/presigned?key="+thumbnailPath, *result.Images[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/presigned?key=users/kp_abc123/images/a.jpg", result.Images[0].FullResURL)

	assert.Equal(t, "second", result.Images[1].Title)
	assert.Nil(t, result.Images[1].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/presigned?key=users/kp_abc123/images/b.jpg", result.Images[1].FullResURL)
}

func TestShareUsecase_GetSharedFolder_UnknownToken(t *testing.T) {
	folderShareRepo := newFakeFolderShareRepo()
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{}, &stubShareImageRepo{}, &spyShareStorage{})

	_, err := uc.GetSharedFolder(context.Background(), "unknown-token")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestShareUsecase_GetSharedFolder_EmptyFolder(t *testing.T) {
	folderID := uuid.New()
	share := &domain.FolderShare{
		ID:       uuid.New(),
		FolderID: folderID,
		Token:    "tok_abc123",
		Folder:   domain.Folder{ID: folderID, UserID: "kp_abc123", Name: "empty"},
	}
	folderShareRepo := newFakeFolderShareRepo(share)
	uc := newTestShareUsecase(folderShareRepo, &stubShareFolderRepo{}, &stubShareImageRepo{images: []*domain.Image{}}, &spyShareStorage{})

	result, err := uc.GetSharedFolder(context.Background(), "tok_abc123")

	require.NoError(t, err)
	assert.Empty(t, result.Images)
}
