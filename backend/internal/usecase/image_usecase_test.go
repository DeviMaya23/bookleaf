package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- mocks ---

type mockImageRepository struct {
	image  *domain.Image
	images []*domain.Image
	err    error
}

func (m *mockImageRepository) Create(_ context.Context, _ *domain.Image) (*domain.Image, error) {
	return m.image, m.err
}

func (m *mockImageRepository) List(_ context.Context, _ string, _ *uuid.UUID) ([]*domain.Image, error) {
	return m.images, m.err
}

func (m *mockImageRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}

func (m *mockImageRepository) GetDeletedByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}

func (m *mockImageRepository) UpdateThumbnailPath(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageRepository) SoftDelete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageRepository) Restore(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageRepository) ListTrashed(_ context.Context, _ string) ([]*domain.Image, error) {
	return m.images, m.err
}

type mockStorageService struct {
	putURL string
	getURL string
	err    error
}

func (m *mockStorageService) GeneratePresignedPutURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return m.putURL, m.err
}

func (m *mockStorageService) GeneratePresignedGetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return m.getURL, m.err
}

func (m *mockStorageService) GetObject(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), m.err
}

func (m *mockStorageService) PutObject(_ context.Context, _ string, _ io.Reader, _ string) error {
	return m.err
}

func (m *mockStorageService) Ping(_ context.Context) error {
	return m.err
}

func (m *mockStorageService) CDNUrl(_ string) string {
	return "https://cdn.example.com/test"
}

type mockThumbnailService struct {
	err error
}

func (m *mockThumbnailService) Generate(_ context.Context, _ io.Reader) (io.Reader, error) {
	return strings.NewReader(""), m.err
}

func noopTel() *observability.Telemetry {
	return observability.NewTelemetry(nil, nil, nil)
}

// --- tests ---

func TestImageUsecase_InitiateUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name     string
		title    string
		mimeType string
		repo     *mockImageRepository
		store    *mockStorageService
		wantErr  error
		wantURL  string
	}{
		{
			name:     "creates image record and returns presigned put url",
			title:    "sunset photo",
			mimeType: "image/jpeg",
			repo:     &mockImageRepository{image: &domain.Image{ID: imageID, Title: "sunset photo"}},
			store:    &mockStorageService{putURL: "https://r2.example.com/upload"},
			wantURL:  "https://r2.example.com/upload",
		},
		{
			name:     "returns error for blank title",
			title:    "   ",
			mimeType: "image/jpeg",
			repo:     &mockImageRepository{},
			store:    &mockStorageService{},
			wantErr:  ErrInvalidImageTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, tt.store, &mockThumbnailService{}, noopTel())

			result, err := uc.InitiateUpload(context.Background(), "kp_abc123", tt.title, tt.mimeType, nil, nil)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, result.UploadURL)
			assert.Equal(t, imageID, result.Image.ID)
		})
	}
}

func TestImageUsecase_CompleteUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockImageRepository
		wantErr bool
	}{
		{
			name: "verifies ownership and fires thumbnail goroutine",
			repo: &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.jpg"}},
		},
		{
			name:    "returns error when image not found",
			repo:    &mockImageRepository{err: gorm.ErrRecordNotFound},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, noopTel())

			err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestImageUsecase_ListImages(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockImageRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "returns images for user",
			repo: &mockImageRepository{
				images: []*domain.Image{
					{ID: uuid.New(), Title: "photo 1"},
					{ID: uuid.New(), Title: "photo 2"},
				},
			},
			wantLen: 2,
		},
		{
			name:    "propagates repository error",
			repo:    &mockImageRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, noopTel())

			images, err := uc.ListImages(context.Background(), "kp_abc123", nil)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, images, tt.wantLen)
		})
	}
}

func TestImageUsecase_GetImage(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockImageRepository
		store   *mockStorageService
		wantURL string
		wantErr bool
	}{
		{
			name:    "returns image detail with presigned get url",
			repo:    &mockImageRepository{image: &domain.Image{ID: imageID, Title: "photo"}},
			store:   &mockStorageService{getURL: "https://r2.example.com/view"},
			wantURL: "https://r2.example.com/view",
		},
		{
			name:    "returns error when image not found",
			repo:    &mockImageRepository{err: gorm.ErrRecordNotFound},
			store:   &mockStorageService{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, tt.store, &mockThumbnailService{}, noopTel())

			detail, err := uc.GetImage(context.Background(), imageID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, detail.ImageURL)
			assert.Equal(t, imageID, detail.Image.ID)
		})
	}
}

func TestImageUsecase_SoftDelete(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockImageRepository
		wantErr bool
	}{
		{
			name: "soft deletes image successfully",
			repo: &mockImageRepository{},
		},
		{
			name:    "propagates repository error",
			repo:    &mockImageRepository{err: gorm.ErrRecordNotFound},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, noopTel())

			err := uc.SoftDelete(context.Background(), imageID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestImageUsecase_ListTrashed(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockImageRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "returns trashed images for user",
			repo: &mockImageRepository{
				images: []*domain.Image{
					{ID: uuid.New(), Title: "deleted photo"},
				},
			},
			wantLen: 1,
		},
		{
			name:    "propagates repository error",
			repo:    &mockImageRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, noopTel())

			images, err := uc.ListTrashed(context.Background(), "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, images, tt.wantLen)
		})
	}
}

func TestImageUsecase_Restore(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockImageRepository
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name:   "restores soft-deleted image",
			repo:   &mockImageRepository{image: &domain.Image{ID: imageID, Title: "photo"}},
			wantID: imageID,
		},
		{
			name:    "returns error when deleted image not found",
			repo:    &mockImageRepository{err: gorm.ErrRecordNotFound},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, noopTel())

			image, err := uc.Restore(context.Background(), imageID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, image.ID)
		})
	}
}
