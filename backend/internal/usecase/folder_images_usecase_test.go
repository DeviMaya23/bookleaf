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

// --- spies ---

type stubFolderImagesFolder struct {
	folder *domain.Folder
	err    error
}

func (s *stubFolderImagesFolder) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.Folder, error) {
	return s.folder, s.err
}

type stubFolderImagesImageRepo struct {
	images []*domain.Image
	err    error
}

func (s *stubFolderImagesImageRepo) ListByFolder(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ *string, _ *string) ([]*domain.Image, error) {
	return s.images, s.err
}

type stubFolderImagesStorage struct {
	StorageService
	err error
}

func (s *stubFolderImagesStorage) GeneratePresignedGetURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return "https://cdn.example.com/" + key, nil
}

func newFolderImagesUsecase(folderRepo *stubFolderImagesFolder, imageRepo *stubFolderImagesImageRepo, store StorageService) FolderImagesUsecase {
	return NewFolderImagesUsecase(folderRepo, imageRepo, store, noopTel())
}

// --- ListFolderImages ---

func TestFolderImagesUsecase_ListFolderImages_ReturnsThumbnails(t *testing.T) {
	thumbPath := "thumbs/img1.jpg"
	img := &domain.Image{ID: uuid.New(), ThumbnailPath: &thumbPath}
	uc := newFolderImagesUsecase(
		&stubFolderImagesFolder{folder: &domain.Folder{ID: uuid.New()}},
		&stubFolderImagesImageRepo{images: []*domain.Image{img}},
		&stubFolderImagesStorage{},
	)

	result, err := uc.ListFolderImages(context.Background(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, img.ID, result[0].ImageID)
	assert.Equal(t, "https://cdn.example.com/thumbs/img1.jpg", result[0].ThumbnailURL)
}

func TestFolderImagesUsecase_ListFolderImages_ExcludesNilThumbnailPath(t *testing.T) {
	thumbPath := "thumbs/img1.jpg"
	imgWithThumb := &domain.Image{ID: uuid.New(), ThumbnailPath: &thumbPath}
	imgNoThumb := &domain.Image{ID: uuid.New(), ThumbnailPath: nil}
	uc := newFolderImagesUsecase(
		&stubFolderImagesFolder{folder: &domain.Folder{ID: uuid.New()}},
		&stubFolderImagesImageRepo{images: []*domain.Image{imgWithThumb, imgNoThumb}},
		&stubFolderImagesStorage{},
	)

	result, err := uc.ListFolderImages(context.Background(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, imgWithThumb.ID, result[0].ImageID)
}

func TestFolderImagesUsecase_ListFolderImages_EmptyFolderReturnsEmptySlice(t *testing.T) {
	uc := newFolderImagesUsecase(
		&stubFolderImagesFolder{folder: &domain.Folder{ID: uuid.New()}},
		&stubFolderImagesImageRepo{images: []*domain.Image{}},
		&stubFolderImagesStorage{},
	)

	result, err := uc.ListFolderImages(context.Background(), uuid.New(), uuid.New())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFolderImagesUsecase_ListFolderImages_UnownedFolderReturnsNotFound(t *testing.T) {
	uc := newFolderImagesUsecase(
		&stubFolderImagesFolder{err: gorm.ErrRecordNotFound},
		&stubFolderImagesImageRepo{},
		&stubFolderImagesStorage{},
	)

	result, err := uc.ListFolderImages(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
