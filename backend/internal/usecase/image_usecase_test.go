package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/vision"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"gorm.io/gorm"
)

// --- mocks ---

type mockImageRepository struct {
	image            *domain.Image
	images           []*domain.Image
	err              error
	count            int64
	updateFields     map[string]any
	lastUpdateID     uuid.UUID
	lastUpdateBy     string
	createdImage     *domain.Image
	lastUnfiled      bool
	hardDeleteCalls  int
}

func (m *mockImageRepository) Create(_ context.Context, image *domain.Image) (*domain.Image, error) {
	m.createdImage = image
	return m.image, m.err
}

func (m *mockImageRepository) List(_ context.Context, _ string, _ *uuid.UUID, unfiled bool, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	m.lastUnfiled = unfiled
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

func (m *mockImageRepository) UpdateAILabels(_ context.Context, _ uuid.UUID, _ json.RawMessage) error {
	return m.err
}

func (m *mockImageRepository) Update(_ context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error) {
	m.lastUpdateID = id
	m.lastUpdateBy = userID
	m.updateFields = _mapCopy(fields)
	return m.image, m.err
}

func (m *mockImageRepository) SoftDelete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageRepository) Restore(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageRepository) ListTrashed(_ context.Context, _ string, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	return m.images, m.err
}

func (m *mockImageRepository) CountByFolderID(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.count, m.err
}

func (m *mockImageRepository) ListStaleUploads(_ context.Context, _ time.Time) ([]*domain.Image, error) {
	return m.images, m.err
}

func (m *mockImageRepository) ListExpiredTrash(_ context.Context, _ time.Time) ([]*domain.Image, error) {
	return m.images, m.err
}

func (m *mockImageRepository) HardDelete(_ context.Context, _ uuid.UUID, _ string) error {
	m.hardDeleteCalls++
	return m.err
}

func _mapCopy(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return out
}

func generateTestPNGBytes(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

type mockStorageService struct {
	putURL          string
	getURL          string
	err             error
	getObjectErr    error
	putObjectErr    error
	deleteObjectErr error
	objectBytes     []byte
	getCalls        int
	putCalls        int
	deleteCalls     int
	deletedKeys     []string
}

func (m *mockStorageService) GeneratePresignedPutURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return m.putURL, m.err
}

func (m *mockStorageService) GeneratePresignedGetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return m.getURL, m.err
}

func (m *mockStorageService) GetObject(_ context.Context, _ string) (io.ReadCloser, error) {
	m.getCalls++
	if m.getObjectErr != nil {
		return nil, m.getObjectErr
	}
	if m.objectBytes != nil {
		return io.NopCloser(bytes.NewReader(m.objectBytes)), m.err
	}
	return io.NopCloser(strings.NewReader("")), m.err
}

func (m *mockStorageService) PutObject(_ context.Context, _ string, _ io.Reader, _ string) error {
	m.putCalls++
	if m.putObjectErr != nil {
		return m.putObjectErr
	}
	return m.err
}

func (m *mockStorageService) DeleteObject(_ context.Context, key string) error {
	m.deleteCalls++
	m.deletedKeys = append(m.deletedKeys, key)
	if m.deleteObjectErr != nil {
		return m.deleteObjectErr
	}
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

type mockImageUserRepository struct {
	user *domain.User
	err  error
}

func (m *mockImageUserRepository) GetOrCreate(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockImageUserRepository) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func defaultMockUserRepo() *mockImageUserRepository {
	return &mockImageUserRepository{user: &domain.User{ID: "kp_abc123"}}
}

type mockVisionService struct {
	labels []visionLabel
	err    error
	calls  int
}

type visionLabel struct {
	Description string
	Score       float32
}

func (m *mockVisionService) AnnotateImage(_ context.Context, _ []byte) ([]vision.Label, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	out := make([]vision.Label, len(m.labels))
	for i, l := range m.labels {
		out[i] = vision.Label{Description: l.Description, Score: l.Score}
	}
	return out, nil
}

type mockAcceptSuggestionFolderRepository struct {
	findResult      *domain.Folder
	findErr         error
	createResult    *domain.Folder
	createErr       error
	findCalls       int
	createCalls     int
	lastFindUserID  string
	lastFindName    string
	lastCreatedData *domain.Folder
}

func (m *mockAcceptSuggestionFolderRepository) Create(_ context.Context, folder *domain.Folder) (*domain.Folder, error) {
	m.createCalls++
	m.lastCreatedData = folder
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.createResult != nil {
		return m.createResult, nil
	}
	return folder, nil
}

func (m *mockAcceptSuggestionFolderRepository) List(_ context.Context, _ string) ([]*domain.Folder, error) {
	return nil, nil
}

func (m *mockAcceptSuggestionFolderRepository) FindByName(_ context.Context, userID, name string) (*domain.Folder, error) {
	m.findCalls++
	m.lastFindUserID = userID
	m.lastFindName = name
	return m.findResult, m.findErr
}

func (m *mockAcceptSuggestionFolderRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Folder, error) {
	return nil, nil
}

func (m *mockAcceptSuggestionFolderRepository) Update(_ context.Context, _ *domain.Folder) (*domain.Folder, error) {
	return nil, nil
}

func (m *mockAcceptSuggestionFolderRepository) CountImagesByFolder(_ context.Context, _ uuid.UUID, _ string) (int, error) {
	return 0, nil
}

func (m *mockAcceptSuggestionFolderRepository) DeleteWithCascade(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func noopTel() *observability.Telemetry {
	return observability.NewTelemetry(nil, nil, nil)
}

func makeMetricsTel(t *testing.T) (*observability.Telemetry, func() metricdata.ResourceMetrics) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
	tel := observability.NewTelemetry(nil, nil, mp.Meter("test"))
	collect := func() metricdata.ResourceMetrics {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &rm))
		return rm
	}
	return tel, collect
}

func findInt64Sum(rm metricdata.ResourceMetrics, name string) []metricdata.DataPoint[int64] {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				if data, ok := m.Data.(metricdata.Sum[int64]); ok {
					return data.DataPoints
				}
			}
		}
	}
	return nil
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
			uc := NewImageUsecase(tt.repo, tt.store, &mockThumbnailService{}, nil, nil, nil, noopTel())

			result, err := uc.InitiateUpload(context.Background(), "kp_abc123", tt.title, tt.mimeType, nil, nil, nil)

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

func TestImageUsecase_InitiateUpload_WithDescription(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := NewImageUsecase(repo, &mockStorageService{putURL: "https://r2.example.com/upload"}, &mockThumbnailService{}, nil, nil, nil, noopTel())
	description := "cover image"

	result, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset photo", "image/jpeg", nil, nil, &description)

	require.NoError(t, err)
	require.NotNil(t, repo.createdImage)
	require.NotNil(t, repo.createdImage.Description)
	assert.Equal(t, description, *repo.createdImage.Description)
	assert.Equal(t, imageID, result.Image.ID)
}

func TestImageUsecase_InitiateUpload_WithFolderValidation(t *testing.T) {
	imageID := uuid.New()
	validFolderID := uuid.New()
	invalidFolderID := uuid.New()

	t.Run("keeps folder_id when folder exists for user", func(t *testing.T) {
		repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
		folderRepo := &mockFolderRepository{
			folder: &domain.Folder{ID: validFolderID, UserID: "kp_abc123", Name: "Nature"},
		}
		uc := NewImageUsecase(
			repo,
			&mockStorageService{putURL: "https://r2.example.com/upload"},
			&mockThumbnailService{},
			nil,
			folderRepo,
			nil,
			noopTel(),
		)

		_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset photo", "image/jpeg", nil, &validFolderID, nil)

		require.NoError(t, err)
		require.NotNil(t, repo.createdImage)
		require.NotNil(t, repo.createdImage.FolderID)
		assert.Equal(t, validFolderID, *repo.createdImage.FolderID)
	})

	t.Run("nulls folder_id when folder is not found", func(t *testing.T) {
		repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
		folderRepo := &mockFolderRepository{
			err: gorm.ErrRecordNotFound,
		}
		uc := NewImageUsecase(
			repo,
			&mockStorageService{putURL: "https://r2.example.com/upload"},
			&mockThumbnailService{},
			nil,
			folderRepo,
			nil,
			noopTel(),
		)

		_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset photo", "image/jpeg", nil, &invalidFolderID, nil)

		require.NoError(t, err)
		require.NotNil(t, repo.createdImage)
		assert.Nil(t, repo.createdImage.FolderID)
	})
}

func TestImageUsecase_CompleteUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name         string
		repo         *mockImageRepository
		store        *mockStorageService
		thumbnails   *mockThumbnailService
		wantErr      bool
		wantWarning  bool
		wantPutCalls *int
	}{
		{
			name:       "verifies ownership and returns upload result",
			repo:       &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.jpg"}},
			store:      &mockStorageService{},
			thumbnails: &mockThumbnailService{},
		},
		{
			name:         "get object failure sets warning and skips goroutine",
			repo:         &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.jpg"}},
			store:        &mockStorageService{getObjectErr: errors.New("r2 unavailable")},
			thumbnails:   &mockThumbnailService{},
			wantWarning:  true,
			wantPutCalls: func() *int { v := 0; return &v }(),
		},
		{
			name:       "returns error when image not found",
			repo:       &mockImageRepository{err: gorm.ErrRecordNotFound},
			store:      &mockStorageService{},
			thumbnails: &mockThumbnailService{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, tt.store, tt.thumbnails, nil, nil, defaultMockUserRepo(), noopTel())

			result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, imageID, result.ImageID)
			if tt.wantWarning {
				assert.NotEmpty(t, result.Warning)
			} else {
				assert.Empty(t, result.Warning)
			}
			if tt.wantPutCalls != nil {
				assert.Equal(t, *tt.wantPutCalls, tt.store.putCalls)
			}
		})
	}
}

func TestImageUsecase_CompleteUpload_PersistsMetadata(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{
		image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.png"},
	}
	store := &mockStorageService{objectBytes: generateTestPNGBytes(t, 8, 6)}
	uc := NewImageUsecase(repo, store, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, repo.updateFields)
	assert.EqualValues(t, len(store.objectBytes), repo.updateFields["file_size"])
	assert.Equal(t, 8, repo.updateFields["width"])
	assert.Equal(t, 6, repo.updateFields["height"])
	assert.Equal(t, true, repo.updateFields["is_uploaded"])
}

func TestImageUsecase_CompleteUpload_DecodeFailureStillPersistsFileSize(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{
		image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.bin"},
	}
	store := &mockStorageService{objectBytes: []byte("not-an-image")}
	uc := NewImageUsecase(repo, store, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, repo.updateFields)
	assert.EqualValues(t, len(store.objectBytes), repo.updateFields["file_size"])
	assert.Nil(t, repo.updateFields["width"])
	assert.Nil(t, repo.updateFields["height"])
}

func TestImageUsecase_CompleteUpload_VisionFlow(t *testing.T) {
	imageID := uuid.New()
	baseImage := &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.jpg"}

	t.Run("vision enabled returns suggested folder name", func(t *testing.T) {
		visionSvc := &mockVisionService{labels: []visionLabel{{Description: "Nature", Score: 0.98}}}
		userRepo := &mockImageUserRepository{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}

		uc := NewImageUsecase(
			&mockImageRepository{image: baseImage},
			&mockStorageService{},
			&mockThumbnailService{},
			visionSvc, nil, userRepo, noopTel(),
		)

		result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

		require.NoError(t, err)
		require.NotNil(t, result.SuggestedFolderName)
		assert.Equal(t, "Nature", *result.SuggestedFolderName)
		assert.Empty(t, result.Warning)
		assert.Equal(t, 1, visionSvc.calls)
	})

	t.Run("vision disabled returns nil suggestion without calling vision service", func(t *testing.T) {
		visionSvc := &mockVisionService{labels: []visionLabel{{Description: "Nature", Score: 0.98}}}
		userRepo := &mockImageUserRepository{user: &domain.User{ID: "kp_abc123", VisionEnabled: false}}

		uc := NewImageUsecase(
			&mockImageRepository{image: baseImage},
			&mockStorageService{},
			&mockThumbnailService{},
			visionSvc, &mockFolderRepository{}, userRepo, noopTel(),
		)

		result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

		require.NoError(t, err)
		assert.Nil(t, result.SuggestedFolderName)
		assert.Empty(t, result.Warning)
		assert.Equal(t, 0, visionSvc.calls)
	})

	t.Run("vision api failure sets warning and returns nil suggestion", func(t *testing.T) {
		visionSvc := &mockVisionService{err: errors.New("vision unavailable")}
		userRepo := &mockImageUserRepository{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}

		uc := NewImageUsecase(
			&mockImageRepository{image: baseImage},
			&mockStorageService{},
			&mockThumbnailService{},
			visionSvc, &mockFolderRepository{}, userRepo, noopTel(),
		)

		result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

		require.NoError(t, err)
		assert.Nil(t, result.SuggestedFolderName)
		assert.NotEmpty(t, result.Warning)
	})
}

func TestImageUsecase_AcceptSuggestion(t *testing.T) {
	imageID := uuid.New()
	existingFolderID := uuid.New()
	newFolderID := uuid.New()

	t.Run("uses existing folder when matched", func(t *testing.T) {
		imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123"}}
		folderRepo := &mockAcceptSuggestionFolderRepository{
			findResult: &domain.Folder{ID: existingFolderID, Name: "Nature"},
		}
		uc := NewImageUsecase(imageRepo, &mockStorageService{}, &mockThumbnailService{}, nil, folderRepo, nil, noopTel())

		err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

		require.NoError(t, err)
		assert.Equal(t, 1, folderRepo.findCalls)
		assert.Equal(t, 0, folderRepo.createCalls)
		require.NotNil(t, imageRepo.updateFields)
		assert.Equal(t, existingFolderID, imageRepo.updateFields["folder_id"])
	})

	t.Run("creates folder when no match is found", func(t *testing.T) {
		imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123"}}
		folderRepo := &mockAcceptSuggestionFolderRepository{
			createResult: &domain.Folder{ID: newFolderID, Name: "Nature"},
		}
		uc := NewImageUsecase(imageRepo, &mockStorageService{}, &mockThumbnailService{}, nil, folderRepo, nil, noopTel())

		err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

		require.NoError(t, err)
		assert.Equal(t, 1, folderRepo.findCalls)
		assert.Equal(t, 1, folderRepo.createCalls)
		require.NotNil(t, folderRepo.lastCreatedData)
		assert.Equal(t, "kp_abc123", folderRepo.lastCreatedData.UserID)
		assert.Equal(t, "Nature", folderRepo.lastCreatedData.Name)
		require.NotNil(t, imageRepo.updateFields)
		assert.Equal(t, newFolderID, imageRepo.updateFields["folder_id"])
	})

	t.Run("returns error when image is not found", func(t *testing.T) {
		imageRepo := &mockImageRepository{err: gorm.ErrRecordNotFound}
		folderRepo := &mockAcceptSuggestionFolderRepository{}
		uc := NewImageUsecase(imageRepo, &mockStorageService{}, &mockThumbnailService{}, nil, folderRepo, nil, noopTel())

		err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Equal(t, 0, folderRepo.findCalls)
		assert.Equal(t, 0, folderRepo.createCalls)
		assert.Nil(t, imageRepo.updateFields)
	})
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
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

			result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{})

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, result.Images, tt.wantLen)
		})
	}
}

func TestImageUsecase_ListImages_Pagination(t *testing.T) {
	now := time.Now().UTC()
	makeImages := func(n int) []*domain.Image {
		imgs := make([]*domain.Image, n)
		for i := range imgs {
			imgs[i] = &domain.Image{ID: uuid.New(), Title: "photo", CreatedAt: now.Add(-time.Duration(i) * time.Second)}
		}
		return imgs
	}

	t.Run("returns next_cursor when more results exist", func(t *testing.T) {
		// repo returns limit+1 rows
		repo := &mockImageRepository{images: makeImages(11)}
		uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

		result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result.Images, 10)
		assert.NotNil(t, result.NextCursor)
		assert.Equal(t, result.Images[9].ID, result.NextCursor.ID)
	})

	t.Run("returns nil next_cursor on last page", func(t *testing.T) {
		repo := &mockImageRepository{images: makeImages(5)}
		uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

		result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result.Images, 5)
		assert.Nil(t, result.NextCursor)
	})
}

func TestImageUsecase_ListImages_Unfiled(t *testing.T) {
	t.Run("Unfiled true passes unfiled flag to repo", func(t *testing.T) {
		repo := &mockImageRepository{
			images: []*domain.Image{{ID: uuid.New(), Title: "unfoldered photo"}},
		}
		uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

		result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Unfiled: true})

		require.NoError(t, err)
		assert.True(t, repo.lastUnfiled)
		assert.Len(t, result.Images, 1)
	})

	t.Run("Unfiled false passes no unfiled flag to repo", func(t *testing.T) {
		repo := &mockImageRepository{
			images: []*domain.Image{{ID: uuid.New(), Title: "any photo"}},
		}
		uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

		result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Unfiled: false})

		require.NoError(t, err)
		assert.False(t, repo.lastUnfiled)
		assert.Len(t, result.Images, 1)
	})
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
			uc := NewImageUsecase(tt.repo, tt.store, &mockThumbnailService{}, nil, nil, nil, noopTel())

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
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

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
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

			result, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{})

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, result.Images, tt.wantLen)
		})
	}
}

func TestImageUsecase_ListTrashed_Pagination(t *testing.T) {
	now := time.Now().UTC()
	makeImages := func(n int) []*domain.Image {
		imgs := make([]*domain.Image, n)
		for i := range imgs {
			imgs[i] = &domain.Image{ID: uuid.New(), Title: "deleted photo", CreatedAt: now.Add(-time.Duration(i) * time.Second)}
		}
		return imgs
	}

	t.Run("returns next_cursor when more results exist", func(t *testing.T) {
		repo := &mockImageRepository{images: makeImages(11)}
		uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

		result, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result.Images, 10)
		assert.NotNil(t, result.NextCursor)
		assert.Equal(t, result.Images[9].ID, result.NextCursor.ID)
	})

	t.Run("returns nil next_cursor on last page", func(t *testing.T) {
		repo := &mockImageRepository{images: makeImages(3)}
		uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

		result, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{Limit: 10})

		require.NoError(t, err)
		assert.Len(t, result.Images, 3)
		assert.Nil(t, result.NextCursor)
	})
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
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

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

func TestImageUsecase_UpdateImage(t *testing.T) {
	imageID := uuid.New()
	title := "new title"

	tests := []struct {
		name    string
		repo    *mockImageRepository
		params  UpdateImageParams
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name: "returns updated image on success",
			repo: &mockImageRepository{image: &domain.Image{ID: imageID, Title: title}},
			params: UpdateImageParams{
				Title: &title,
			},
			wantID: imageID,
		},
		{
			name:    "propagates repository error",
			repo:    &mockImageRepository{err: gorm.ErrRecordNotFound},
			params:  UpdateImageParams{Title: &title},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewImageUsecase(tt.repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())

			image, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", tt.params)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, image.ID)
		})
	}
}

func TestImageUsecase_UpdateImage_WithDescription(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, noopTel())
	description := "new description"

	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{
		Description: &description,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.updateFields)
	assert.Equal(t, description, repo.updateFields["description"])
}

func TestImageUsecase_CompleteUpload_UploadCount_Success(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	repo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.jpg"}}

	uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), tel)
	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")
	require.NoError(t, err)

	rm := collect()
	points := findInt64Sum(rm, "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)

	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "success", status.AsString())
}

func TestImageUsecase_CompleteUpload_UploadCount_Error(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	repo := &mockImageRepository{err: gorm.ErrRecordNotFound}

	uc := NewImageUsecase(repo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, tel)
	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")
	require.Error(t, err)

	rm := collect()
	points := findInt64Sum(rm, "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)

	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "error", status.AsString())
}

func TestImageUsecase_CleanupStaleUploads_Success(t *testing.T) {
	imgID1 := uuid.New()
	imgID2 := uuid.New()
	staleImages := []*domain.Image{
		{ID: imgID1, UserID: "kp_user1", R2Path: "users/kp_user1/images/a.jpg"},
		{ID: imgID2, UserID: "kp_user2", R2Path: "users/kp_user2/images/b.jpg"},
	}
	repo := &mockImageRepository{images: staleImages}
	store := &mockStorageService{}
	uc := NewImageUsecase(repo, store, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), noopTel())

	err := uc.CleanupStaleUploads(context.Background(), 30*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_user1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, "users/kp_user2/images/b.jpg")
	assert.Equal(t, 2, repo.hardDeleteCalls)
}

func TestImageUsecase_CleanupStaleUploads_ListError(t *testing.T) {
	repo := &mockImageRepository{err: errors.New("db unavailable")}
	store := &mockStorageService{}
	uc := NewImageUsecase(repo, store, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), noopTel())

	err := uc.CleanupStaleUploads(context.Background(), 30*time.Minute)

	require.Error(t, err)
	assert.Equal(t, 0, store.deleteCalls)
}

func TestImageUsecase_PurgeExpiredTrash_Success(t *testing.T) {
	thumbnailPath := "users/kp_user1/thumbnails/b.jpg"
	expiredImages := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_user1", R2Path: "users/kp_user1/images/a.jpg"},
		{ID: uuid.New(), UserID: "kp_user1", R2Path: "users/kp_user1/images/b.jpg", ThumbnailPath: &thumbnailPath},
	}
	repo := &mockImageRepository{images: expiredImages}
	store := &mockStorageService{}
	uc := NewImageUsecase(repo, store, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), noopTel())

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.NoError(t, err)
	// 2 r2_path deletes + 1 thumbnail delete
	assert.Equal(t, 3, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_user1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, "users/kp_user1/images/b.jpg")
	assert.Contains(t, store.deletedKeys, thumbnailPath)
}

func TestImageUsecase_PurgeExpiredTrash_ListError(t *testing.T) {
	repo := &mockImageRepository{err: errors.New("db unavailable")}
	store := &mockStorageService{}
	uc := NewImageUsecase(repo, store, &mockThumbnailService{}, nil, nil, defaultMockUserRepo(), noopTel())

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.Error(t, err)
	assert.Equal(t, 0, store.deleteCalls)
}
