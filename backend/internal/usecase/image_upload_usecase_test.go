package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"gorm.io/gorm"
)

// --- test doubles ---

type mockPendingUploadRepository struct {
	pending              *domain.PendingUpload
	pendings             []*domain.PendingUpload
	err                  error
	createErr            error
	getErr               error
	transactionCalls     int
	transactionImageRepo ImageRepository
	createdPending       *domain.PendingUpload
	deletedID            uuid.UUID
	deleteCalls          int
}

func (m *mockPendingUploadRepository) Create(_ context.Context, pending *domain.PendingUpload) (*domain.PendingUpload, error) {
	m.createdPending = pending
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.pending != nil {
		return m.pending, m.err
	}
	return pending, m.err
}
func (m *mockPendingUploadRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.PendingUpload, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.pending, m.err
}
func (m *mockPendingUploadRepository) Delete(_ context.Context, id uuid.UUID) error {
	m.deleteCalls++
	m.deletedID = id
	return m.err
}
func (m *mockPendingUploadRepository) ListStale(_ context.Context, _ time.Time) ([]*domain.PendingUpload, error) {
	return m.pendings, m.err
}
func (m *mockPendingUploadRepository) Transaction(_ context.Context, fn func(PendingUploadRepository, ImageRepository) error) error {
	m.transactionCalls++
	imageRepo := m.transactionImageRepo
	if imageRepo == nil {
		imageRepo = &mockImageRepository{}
	}
	return fn(m, imageRepo)
}

type stubUserRepo struct{ user *domain.User }

func (s *stubUserRepo) GetOrCreate(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}

type stubImageFolderRepo struct {
	folder *domain.Folder
	err    error
}

func (s *stubImageFolderRepo) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Folder, error) {
	return s.folder, s.err
}
func (s *stubImageFolderRepo) FindByName(_ context.Context, _, _ string) (*domain.Folder, error) {
	return s.folder, s.err
}
func (s *stubImageFolderRepo) Create(_ context.Context, folder *domain.Folder) (*domain.Folder, error) {
	return folder, s.err
}

type mockVisionService struct {
	labels []domain.Label
	err    error
	calls  int
}

func (m *mockVisionService) AnnotateImage(_ context.Context, _ []byte) ([]domain.Label, error) {
	m.calls++
	return m.labels, m.err
}

// --- helpers ---

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

func defaultUserRepo() *stubUserRepo {
	return &stubUserRepo{user: &domain.User{ID: "kp_abc123"}}
}

type noopJobEnqueuer struct{}

func (n *noopJobEnqueuer) Insert(_ context.Context, _ JobArgs) error { return nil }

type spyJobEnqueuer struct {
	insertCalls int
}

func (s *spyJobEnqueuer) Insert(_ context.Context, _ JobArgs) error {
	s.insertCalls++
	return nil
}

func newImageUploadUsecase(
	imageRepo UploadImageRepository,
	pendingRepo PendingUploadRepository,
	folderRepo ImageFolderRepository,
	userRepo UserRepository,
	store StorageService,
	vision VisionService,
) *imageUploadUsecase {
	return NewImageUploadUsecase(imageRepo, pendingRepo, folderRepo, userRepo, store, vision, &noopJobEnqueuer{}, noopTel())
}

// --- InitiateUpload ---

func TestImageUploadUsecase_InitiateUpload_BlankTitle(t *testing.T) {
	uc := newImageUploadUsecase(&mockImageRepository{}, &mockPendingUploadRepository{}, nil, nil, &mockStorageService{}, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "   ", "image/jpeg", nil, nil, nil)

	require.ErrorIs(t, err, ErrInvalidImageTitle)
}

func TestImageUploadUsecase_InitiateUpload_BlankMimeType(t *testing.T) {
	uc := newImageUploadUsecase(&mockImageRepository{}, &mockPendingUploadRepository{}, nil, nil, &mockStorageService{}, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "   ", nil, nil, nil)

	require.ErrorIs(t, err, ErrInvalidMIMEType)
}

func TestImageUploadUsecase_InitiateUpload_R2PathFormat(t *testing.T) {
	pendingRepo := &mockPendingUploadRepository{}
	uc := newImageUploadUsecase(&mockImageRepository{}, pendingRepo, nil, nil, &mockStorageService{putURL: "https://r2.example.com/upload"}, nil)

	result, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, pendingRepo.createdPending)
	assert.True(t, strings.HasPrefix(result.R2Path, "users/kp_abc123/images/"))
	assert.True(t, strings.HasSuffix(result.R2Path, ".jpg"))
	assert.Equal(t, result.ID, pendingRepo.createdPending.ID)
	assert.NotEmpty(t, result.ThumbnailUploadURL)
	assert.True(t, strings.HasPrefix(result.ThumbnailKey, "users/kp_abc123/thumbnails/"))
	assert.True(t, strings.HasSuffix(result.ThumbnailKey, ".jpg"))
}

func TestImageUploadUsecase_InitiateUpload_FolderFound(t *testing.T) {
	folderID := uuid.New()
	pendingRepo := &mockPendingUploadRepository{}
	folderRepo := newFakeFolderRepo(&domain.Folder{ID: folderID, UserID: "kp_abc123"})
	uc := newImageUploadUsecase(&mockImageRepository{}, pendingRepo, folderRepo, nil, &mockStorageService{putURL: "https://r2.example.com/upload"}, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, &folderID, nil)

	require.NoError(t, err)
	require.NotNil(t, pendingRepo.createdPending.FolderID)
	assert.Equal(t, folderID, *pendingRepo.createdPending.FolderID)
}

func TestImageUploadUsecase_InitiateUpload_FolderNotFound(t *testing.T) {
	folderID := uuid.New()
	pendingRepo := &mockPendingUploadRepository{}
	folderRepo := &stubImageFolderRepo{err: gorm.ErrRecordNotFound}
	uc := newImageUploadUsecase(&mockImageRepository{}, pendingRepo, folderRepo, nil, &mockStorageService{putURL: "https://r2.example.com/upload"}, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, &folderID, nil)

	require.NoError(t, err)
	assert.Nil(t, pendingRepo.createdPending.FolderID)
}

func TestImageUploadUsecase_InitiateUpload_CreatePendingFails(t *testing.T) {
	pendingRepo := &mockPendingUploadRepository{createErr: errors.New("db error")}
	uc := newImageUploadUsecase(&mockImageRepository{}, pendingRepo, nil, nil, &mockStorageService{}, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, nil, nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, "create pending upload record")
}

// --- CompleteUpload ---

func TestImageUploadUsecase_CompleteUpload_PersistsImageMetadata(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.png", MIMEType: "image/png"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	store := &mockStorageService{objectBytes: generateTestPNGBytes(t, 8, 6)}
	uc := NewImageUploadUsecase(imageRepo, pendingRepo, nil, defaultUserRepo(), store, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.createdImage)
	require.NotNil(t, imageRepo.createdImage.FileSize)
	assert.EqualValues(t, len(store.objectBytes), *imageRepo.createdImage.FileSize)
	require.NotNil(t, imageRepo.createdImage.Width)
	require.NotNil(t, imageRepo.createdImage.Height)
	assert.Equal(t, 8, *imageRepo.createdImage.Width)
	assert.Equal(t, 6, *imageRepo.createdImage.Height)
}

func TestImageUploadUsecase_CompleteUpload_DecodeFailurePersistsFileSize(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.bin", MIMEType: "application/octet-stream"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	store := &mockStorageService{objectBytes: []byte("not-an-image")}
	uc := NewImageUploadUsecase(imageRepo, pendingRepo, nil, defaultUserRepo(), store, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.createdImage.FileSize)
	assert.EqualValues(t, len(store.objectBytes), *imageRepo.createdImage.FileSize)
	assert.Nil(t, imageRepo.createdImage.Width)
	assert.Nil(t, imageRepo.createdImage.Height)
}

func TestImageUploadUsecase_CompleteUpload_SetsFolderWhenPendingHasFolderID(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg", FolderID: &folderID},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderID, *imageRepo.setFolderFolderID)
}

func TestImageUploadUsecase_CompleteUpload_AlwaysSetsThumbnailPath(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	enqueuer := &spyJobEnqueuer{}
	uc := NewImageUploadUsecase(imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, enqueuer, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.createdImage.ThumbnailPath)
	assert.True(t, strings.HasPrefix(*imageRepo.createdImage.ThumbnailPath, "users/kp_abc123/thumbnails/"))
	assert.Equal(t, 1, enqueuer.insertCalls, "only vision job should be enqueued")
}

// --- ProcessVisionLabelling ---

func TestImageUploadUsecase_ProcessVisionLabelling_SavesLabels(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/img.jpg"}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{labels: []domain.Label{{Description: "Nature", Score: 0.98}}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc)

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 1, visionSvc.calls)
	assert.Equal(t, 1, imageRepo.updateAILabelsCalls)
	var saved []domain.Label
	require.NoError(t, json.Unmarshal(imageRepo.lastAILabels, &saved))
	require.Len(t, saved, 1)
	assert.Equal(t, "Nature", saved[0].Description)
}

func TestImageUploadUsecase_ProcessVisionLabelling_ZeroLabelsWritesEmpty(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/img.jpg"}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{labels: []domain.Label{}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc)

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 1, imageRepo.updateAILabelsCalls)
	assert.Equal(t, "[]", string(imageRepo.lastAILabels))
}

func TestImageUploadUsecase_ProcessVisionLabelling_VisionDisabledReturnsNil(t *testing.T) {
	imageID := uuid.New()
	visionSvc := &mockVisionService{}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: false}}
	uc := newImageUploadUsecase(&mockImageRepository{}, &mockPendingUploadRepository{}, nil, userRepo, &mockStorageService{}, visionSvc)

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 0, visionSvc.calls)
}

func TestImageUploadUsecase_ProcessVisionLabelling_VisionAPIErrorReturnsError(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/img.jpg"}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{err: errors.New("vision API unavailable")}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc)

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.Error(t, err)
	assert.ErrorContains(t, err, "annotate image")
}

func TestImageUploadUsecase_CompleteUpload_UploadCount_Success(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, tel)

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")
	require.NoError(t, err)

	points := findInt64Sum(collect(), "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)
	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "success", status.AsString())
}

func TestImageUploadUsecase_CompleteUpload_UploadCount_Error(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	pendingRepo := &mockPendingUploadRepository{getErr: gorm.ErrRecordNotFound}
	uc := NewImageUploadUsecase(&mockImageRepository{}, pendingRepo, nil, nil, &mockStorageService{}, nil, &noopJobEnqueuer{}, tel)

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")
	require.Error(t, err)

	points := findInt64Sum(collect(), "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)
	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "error", status.AsString())
}

// --- AcceptSuggestion ---

func TestImageUploadUsecase_AcceptSuggestion_ExistingFolder(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123"}}
	folderRepo := newFakeFolderRepo(&domain.Folder{ID: folderID, Name: "Nature", UserID: "kp_abc123"})
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, folderRepo, nil, &mockStorageService{}, nil)

	err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderID, *imageRepo.setFolderFolderID)
}

func TestImageUploadUsecase_AcceptSuggestion_CreatesFolder(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123"}}
	folderRepo := newFakeFolderRepo()
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, folderRepo, nil, &mockStorageService{}, nil)

	err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

	require.NoError(t, err)
	created, ok := folderRepo.folders["nature"]
	require.True(t, ok)
	assert.Equal(t, "Nature", created.Name)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, created.ID, *imageRepo.setFolderFolderID)
}

// --- CleanupStaleUploads ---

func TestImageUploadUsecase_CleanupStaleUploads_DeletesAll(t *testing.T) {
	stale := []*domain.PendingUpload{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
		{ID: uuid.New(), UserID: "kp_u2", R2Path: "users/kp_u2/images/b.jpg"},
	}
	pendingRepo := &mockPendingUploadRepository{pendings: stale}
	store := &mockStorageService{}
	uc := newImageUploadUsecase(&mockImageRepository{}, pendingRepo, nil, nil, store, nil)

	err := uc.CleanupStaleUploads(context.Background(), 30*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_u1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, "users/kp_u2/images/b.jpg")
	assert.Equal(t, 2, pendingRepo.deleteCalls)
}

func TestImageUploadUsecase_CleanupStaleUploads_StorageErrorContinues(t *testing.T) {
	stale := []*domain.PendingUpload{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
	}
	pendingRepo := &mockPendingUploadRepository{pendings: stale}
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newImageUploadUsecase(&mockImageRepository{}, pendingRepo, nil, nil, store, nil)

	err := uc.CleanupStaleUploads(context.Background(), 30*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 1, pendingRepo.deleteCalls)
}
