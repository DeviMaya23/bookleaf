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

type mockImageRepository struct {
	image             *domain.Image
	images            []*domain.Image
	err               error
	createdImage      *domain.Image
	updateFields      map[string]any
	lastUpdateID      uuid.UUID
	lastUpdateBy      string
	hardDeleteCalls   int
	setFolderCalls    int
	setFolderImageID  uuid.UUID
	setFolderFolderID *uuid.UUID
	syncFolderCalls   int
	lastSyncImageID   uuid.UUID
	lastSyncFolderIDs []uuid.UUID
	moveFolderCalls   int
	lastMoveImageID   uuid.UUID
	lastMoveFromFolderID *uuid.UUID
	lastMoveToFolderID   *uuid.UUID
}

func (m *mockImageRepository) Create(_ context.Context, img *domain.Image) (*domain.Image, error) {
	m.createdImage = img
	return m.image, m.err
}
func (m *mockImageRepository) List(_ context.Context, _ string, _ *uuid.UUID, _ bool, _ *uuid.UUID, _ *ImageCursor, _ int) ([]*domain.Image, error) {
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
	return 0, m.err
}
func (m *mockImageRepository) ListExpiredTrash(_ context.Context, _ time.Time) ([]*domain.Image, error) {
	return m.images, m.err
}
func (m *mockImageRepository) HardDelete(_ context.Context, _ uuid.UUID, _ string) error {
	m.hardDeleteCalls++
	return m.err
}
func (m *mockImageRepository) SetImageFolder(_ context.Context, imageID uuid.UUID, folderID *uuid.UUID) error {
	m.setFolderCalls++
	m.setFolderImageID = imageID
	m.setFolderFolderID = folderID
	return m.err
}
func (m *mockImageRepository) SyncImageFolders(_ context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error {
	m.syncFolderCalls++
	m.lastSyncImageID = imageID
	m.lastSyncFolderIDs = append([]uuid.UUID(nil), folderIDs...)
	return m.err
}
func (m *mockImageRepository) MoveImageFolder(_ context.Context, imageID uuid.UUID, from *uuid.UUID, to *uuid.UUID) error {
	m.moveFolderCalls++
	m.lastMoveImageID = imageID
	m.lastMoveFromFolderID = from
	m.lastMoveToFolderID = to
	return m.err
}
func (m *mockImageRepository) UpdateImageFolderPosition(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return m.err
}

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

type mockStorageService struct {
	putURL      string
	getURL      string
	downloadURL string
	err         error
	getObjectErr  error
	putObjectErr  error
	deleteObjectErr error
	objectBytes []byte
	getCalls    int
	putCalls    int
	deleteCalls int
	deletedKeys []string
	lastDownloadKey      string
	lastDownloadFilename string
	lastDownloadTTL      time.Duration
}

func (m *mockStorageService) GeneratePresignedPutURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return m.putURL, m.err
}
func (m *mockStorageService) GeneratePresignedGetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return m.getURL, m.err
}
func (m *mockStorageService) GeneratePresignedDownloadURL(_ context.Context, key, filename string, ttl time.Duration) (string, error) {
	m.lastDownloadKey = key
	m.lastDownloadFilename = filename
	m.lastDownloadTTL = ttl
	return m.downloadURL, m.err
}
func (m *mockStorageService) GetObject(_ context.Context, _ string) (io.ReadCloser, error) {
	m.getCalls++
	if m.getObjectErr != nil {
		return nil, m.getObjectErr
	}
	if m.objectBytes != nil {
		return io.NopCloser(bytes.NewReader(m.objectBytes)), nil
	}
	return io.NopCloser(strings.NewReader("")), nil
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
func (m *mockStorageService) Ping(_ context.Context) error { return m.err }

type mockThumbnailService struct{ err error }

func (m *mockThumbnailService) Generate(_ context.Context, _ io.Reader) (io.Reader, error) {
	return strings.NewReader(""), m.err
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

type stubUserRepo struct{ user *domain.User }

func (s *stubUserRepo) GetOrCreate(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
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

type mockTagRepository struct {
	tag                *domain.Tag
	tags               []*domain.Tag
	err                error
	replaceCalls       int
	lastReplaceImageID uuid.UUID
	lastReplaceTagIDs  []uuid.UUID
}

func (m *mockTagRepository) Create(_ context.Context, _ *domain.Tag) (*domain.Tag, error) {
	return m.tag, m.err
}
func (m *mockTagRepository) ListByUserID(_ context.Context, _ string) ([]*domain.Tag, error) {
	return m.tags, m.err
}
func (m *mockTagRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Tag, error) {
	return m.tag, m.err
}
func (m *mockTagRepository) Update(_ context.Context, _ uuid.UUID, _ string, _ string) (*domain.Tag, error) {
	return m.tag, m.err
}
func (m *mockTagRepository) Delete(_ context.Context, _ uuid.UUID, _ string) error { return m.err }
func (m *mockTagRepository) ReplaceImageTags(_ context.Context, imageID uuid.UUID, tagIDs []uuid.UUID) error {
	m.replaceCalls++
	m.lastReplaceImageID = imageID
	if tagIDs != nil {
		m.lastReplaceTagIDs = append([]uuid.UUID(nil), tagIDs...)
	} else {
		m.lastReplaceTagIDs = nil
	}
	return m.err
}

// --- helpers ---

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

func defaultUserRepo() *stubUserRepo {
	return &stubUserRepo{user: &domain.User{ID: "kp_abc123"}}
}

func newImageUsecase(
	imageRepo ImageRepository,
	pendingRepo PendingUploadRepository,
	tagRepo TagRepository,
	store StorageService,
	thumbnails ThumbnailService,
	vision VisionService,
	folderRepo ImageFolderRepository,
	userRepo UserRepository,
) *imageUsecase {
	return NewImageUsecase(imageRepo, pendingRepo, tagRepo, store, thumbnails, vision, folderRepo, userRepo, noopTel())
}

// --- InitiateUpload ---

func TestImageUsecase_InitiateUpload_BlankTitle(t *testing.T) {
	uc := newImageUsecase(&mockImageRepository{}, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "   ", "image/jpeg", nil, nil, nil)

	require.ErrorIs(t, err, ErrInvalidImageTitle)
}

func TestImageUsecase_InitiateUpload_BlankMimeType(t *testing.T) {
	uc := newImageUsecase(&mockImageRepository{}, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "   ", nil, nil, nil)

	require.ErrorIs(t, err, ErrInvalidMIMEType)
}

func TestImageUsecase_InitiateUpload_R2PathFormat(t *testing.T) {
	pendingRepo := &mockPendingUploadRepository{}
	uc := newImageUsecase(&mockImageRepository{}, pendingRepo, nil, &mockStorageService{putURL: "https://r2.example.com/upload"}, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, pendingRepo.createdPending)
	assert.True(t, strings.HasPrefix(result.R2Path, "users/kp_abc123/images/"))
	assert.True(t, strings.HasSuffix(result.R2Path, ".jpg"))
	assert.Equal(t, result.ID, pendingRepo.createdPending.ID)
}

func TestImageUsecase_InitiateUpload_FolderFound(t *testing.T) {
	folderID := uuid.New()
	pendingRepo := &mockPendingUploadRepository{}
	folderRepo := &stubImageFolderRepo{folder: &domain.Folder{ID: folderID, UserID: "kp_abc123"}}
	uc := newImageUsecase(&mockImageRepository{}, pendingRepo, nil, &mockStorageService{putURL: "https://r2.example.com/upload"}, &mockThumbnailService{}, nil, folderRepo, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, &folderID, nil)

	require.NoError(t, err)
	require.NotNil(t, pendingRepo.createdPending.FolderID)
	assert.Equal(t, folderID, *pendingRepo.createdPending.FolderID)
}

func TestImageUsecase_InitiateUpload_FolderNotFound(t *testing.T) {
	folderID := uuid.New()
	pendingRepo := &mockPendingUploadRepository{}
	folderRepo := &stubImageFolderRepo{err: gorm.ErrRecordNotFound}
	uc := newImageUsecase(&mockImageRepository{}, pendingRepo, nil, &mockStorageService{putURL: "https://r2.example.com/upload"}, &mockThumbnailService{}, nil, folderRepo, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, &folderID, nil)

	require.NoError(t, err)
	assert.Nil(t, pendingRepo.createdPending.FolderID)
}

func TestImageUsecase_InitiateUpload_CreatePendingFails(t *testing.T) {
	pendingRepo := &mockPendingUploadRepository{createErr: errors.New("db error")}
	uc := newImageUsecase(&mockImageRepository{}, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.InitiateUpload(context.Background(), "kp_abc123", "sunset", "image/jpeg", nil, nil, nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, "create pending upload record")
}

// --- CompleteUpload ---

func TestImageUsecase_CompleteUpload_PersistsImageMetadata(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.png", MIMEType: "image/png"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	store := &mockStorageService{objectBytes: generateTestPNGBytes(t, 8, 6)}
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, store, &mockThumbnailService{}, nil, nil, defaultUserRepo(), noopTel())

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

func TestImageUsecase_CompleteUpload_DecodeFailurePersistsFileSize(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.bin", MIMEType: "application/octet-stream"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	store := &mockStorageService{objectBytes: []byte("not-an-image")}
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, store, &mockThumbnailService{}, nil, nil, defaultUserRepo(), noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.createdImage.FileSize)
	assert.EqualValues(t, len(store.objectBytes), *imageRepo.createdImage.FileSize)
	assert.Nil(t, imageRepo.createdImage.Width)
	assert.Nil(t, imageRepo.createdImage.Height)
}

func TestImageUsecase_CompleteUpload_SetsFolderWhenPendingHasFolderID(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg", FolderID: &folderID},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, defaultUserRepo(), noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderID, *imageRepo.setFolderFolderID)
}

func TestImageUsecase_CompleteUpload_VisionEnabled(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	visionSvc := &mockVisionService{labels: []domain.Label{{Description: "Nature", Score: 0.98}}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, visionSvc, nil, userRepo, noopTel())

	result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.NotNil(t, result.SuggestedFolderName)
	assert.Equal(t, "Nature", *result.SuggestedFolderName)
	assert.Empty(t, result.Warning)
	assert.Equal(t, 1, visionSvc.calls)
}

func TestImageUsecase_CompleteUpload_VisionDisabled(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	visionSvc := &mockVisionService{labels: []domain.Label{{Description: "Nature", Score: 0.98}}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: false}}
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, visionSvc, nil, userRepo, noopTel())

	result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Nil(t, result.SuggestedFolderName)
	assert.Empty(t, result.Warning)
	assert.Equal(t, 0, visionSvc.calls)
}

func TestImageUsecase_CompleteUpload_VisionFails(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	visionSvc := &mockVisionService{err: errors.New("vision unavailable")}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, visionSvc, nil, userRepo, noopTel())

	result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Nil(t, result.SuggestedFolderName)
	assert.NotEmpty(t, result.Warning)
}

func TestImageUsecase_CompleteUpload_UploadCount_Success(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUsecase(imageRepo, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, defaultUserRepo(), tel)

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123")
	require.NoError(t, err)

	points := findInt64Sum(collect(), "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)
	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "success", status.AsString())
}

func TestImageUsecase_CompleteUpload_UploadCount_Error(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	pendingRepo := &mockPendingUploadRepository{getErr: gorm.ErrRecordNotFound}
	uc := NewImageUsecase(&mockImageRepository{}, pendingRepo, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil, tel)

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

func TestImageUsecase_AcceptSuggestion_ExistingFolder(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123"}}
	folderRepo := newFakeFolderRepo(&domain.Folder{ID: folderID, Name: "Nature", UserID: "kp_abc123"})
	uc := newImageUsecase(imageRepo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, folderRepo, nil)

	err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

	require.NoError(t, err)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, folderID, *imageRepo.setFolderFolderID)
}

func TestImageUsecase_AcceptSuggestion_CreatesFolder(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123"}}
	folderRepo := newFakeFolderRepo()
	uc := newImageUsecase(imageRepo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, folderRepo, nil)

	err := uc.AcceptSuggestion(context.Background(), imageID, "kp_abc123", "Nature")

	require.NoError(t, err)
	created, ok := folderRepo.folders["nature"]
	require.True(t, ok)
	assert.Equal(t, "Nature", created.Name)
	require.NotNil(t, imageRepo.setFolderFolderID)
	assert.Equal(t, created.ID, *imageRepo.setFolderFolderID)
}

// --- ListImages ---

func TestImageUsecase_ListImages_FolderView(t *testing.T) {
	folderID := uuid.New()
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}, {ID: uuid.New()}}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{FolderID: &folderID})

	require.NoError(t, err)
	assert.Len(t, result.Images, 2)
	assert.Nil(t, result.NextCursor)
}

func TestImageUsecase_ListImages_NextCursor(t *testing.T) {
	now := time.Now().UTC()
	images := make([]*domain.Image, 11)
	for i := range images {
		images[i] = &domain.Image{ID: uuid.New(), CreatedAt: now.Add(-time.Duration(i) * time.Second)}
	}
	repo := &mockImageRepository{images: images}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, result.Images, 10)
	require.NotNil(t, result.NextCursor)
	assert.Equal(t, result.Images[9].Image.ID, result.NextCursor.ID)
}

func TestImageUsecase_ListImages_LastPage(t *testing.T) {
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}, {ID: uuid.New()}}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, result.Images, 2)
	assert.Nil(t, result.NextCursor)
}

func TestImageUsecase_ListImages_ThumbnailURL(t *testing.T) {
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New(), ThumbnailPath: &thumbnailPath}}}
	store := &mockStorageService{getURL: "https://r2.example.com/thumb"}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{})

	require.NoError(t, err)
	require.Len(t, result.Images, 1)
	require.NotNil(t, result.Images[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/thumb", *result.Images[0].ThumbnailURL)
}

// --- GetImage ---

func TestImageUsecase_GetImage_WithThumbnail(t *testing.T) {
	imageID := uuid.New()
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	repo := &mockImageRepository{image: &domain.Image{ID: imageID, ThumbnailPath: &thumbnailPath}}
	store := &mockStorageService{getURL: "https://r2.example.com/view"}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	detail, err := uc.GetImage(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, imageID, detail.Image.ID)
	assert.Equal(t, "https://r2.example.com/view", detail.ImageURL)
	require.NotNil(t, detail.ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/view", *detail.ThumbnailURL)
}

func TestImageUsecase_GetImage_NoThumbnail(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	store := &mockStorageService{getURL: "https://r2.example.com/view"}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	detail, err := uc.GetImage(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "https://r2.example.com/view", detail.ImageURL)
	assert.Nil(t, detail.ThumbnailURL)
}

func TestImageUsecase_GetImage_PresignFails(t *testing.T) {
	repo := &mockImageRepository{image: &domain.Image{ID: uuid.New()}}
	store := &mockStorageService{err: errors.New("presign failed")}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.GetImage(context.Background(), uuid.New(), "kp_abc123")

	require.Error(t, err)
}

// --- DownloadImage ---

func TestImageUsecase_DownloadImage(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID, Title: "sunset", MIMEType: "image/jpeg", R2Path: "users/kp_abc123/images/photo.jpg"}}
	store := &mockStorageService{downloadURL: "https://r2.example.com/download"}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	url, err := uc.DownloadImage(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "https://r2.example.com/download", url)
	assert.Equal(t, "sunset.jpg", store.lastDownloadFilename)
	assert.Equal(t, downloadURLTTL, store.lastDownloadTTL)
}

// --- ListTrashed ---

func TestImageUsecase_ListTrashed_NextCursor(t *testing.T) {
	now := time.Now().UTC()
	images := make([]*domain.Image, 11)
	for i := range images {
		deletedAt := now.Add(time.Duration(i) * time.Second)
		images[i] = &domain.Image{
			ID:        uuid.New(),
			DeletedAt: gorm.DeletedAt{Time: deletedAt, Valid: true},
		}
	}
	repo := &mockImageRepository{images: images}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, result.Images, 10)
	require.NotNil(t, result.NextCursor)
	assert.Equal(t, result.Images[9].Image.ID, result.NextCursor.ID)
	require.NotNil(t, result.NextCursor.DeletedAt)
	assert.Equal(t, result.Images[9].Image.DeletedAt.Time.UTC(), result.NextCursor.DeletedAt.UTC())
}

func TestImageUsecase_ListTrashed_LastPage(t *testing.T) {
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}, {ID: uuid.New()}}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, result.Images, 2)
	assert.Nil(t, result.NextCursor)
}

func TestImageUsecase_ListTrashed_ThumbnailURL(t *testing.T) {
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New(), ThumbnailPath: &thumbnailPath}}}
	store := &mockStorageService{getURL: "https://r2.example.com/thumb"}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	result, err := uc.ListTrashed(context.Background(), "kp_abc123", ListTrashedParams{})

	require.NoError(t, err)
	require.Len(t, result.Images, 1)
	require.NotNil(t, result.Images[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/thumb", *result.Images[0].ThumbnailURL)
}

// --- Restore ---

func TestImageUsecase_Restore(t *testing.T) {
	imageID := uuid.New()
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	repo := &mockImageRepository{image: &domain.Image{ID: imageID, ThumbnailPath: &thumbnailPath}}
	store := &mockStorageService{getURL: "https://r2.example.com/thumb"}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	item, err := uc.Restore(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, imageID, item.Image.ID)
	require.NotNil(t, item.ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/thumb", *item.ThumbnailURL)
}

// --- UpdateImage ---

func TestImageUsecase_UpdateImage_FieldsAssembled(t *testing.T) {
	imageID := uuid.New()
	title := "new title"
	description := "new description"
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{
		Title:       &title,
		Description: &description,
	})

	require.NoError(t, err)
	assert.Equal(t, title, repo.updateFields["title"])
	assert.Equal(t, description, repo.updateFields["description"])
}

func TestImageUsecase_UpdateImage_NilTags_NoReplace(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	tagRepo := &mockTagRepository{}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, tagRepo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{Tags: nil})

	require.NoError(t, err)
	assert.Equal(t, 0, tagRepo.replaceCalls)
}

func TestImageUsecase_UpdateImage_WithTags(t *testing.T) {
	imageID := uuid.New()
	tag1, tag2 := uuid.New(), uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	tagRepo := &mockTagRepository{}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, tagRepo, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	tags := []uuid.UUID{tag1, tag2}
	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{Tags: &tags})

	require.NoError(t, err)
	assert.Equal(t, 1, tagRepo.replaceCalls)
	assert.Equal(t, imageID, tagRepo.lastReplaceImageID)
	assert.Equal(t, tags, tagRepo.lastReplaceTagIDs)
}

func TestImageUsecase_UpdateImage_NilFolderIDs_NoSync(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{FolderIDs: nil})

	require.NoError(t, err)
	assert.Equal(t, 0, repo.syncFolderCalls)
}

func TestImageUsecase_UpdateImage_WithFolderIDs(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	folderIDs := []uuid.UUID{folderID}
	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{FolderIDs: &folderIDs})

	require.NoError(t, err)
	assert.Equal(t, 1, repo.syncFolderCalls)
	assert.Equal(t, imageID, repo.lastSyncImageID)
	assert.Equal(t, folderIDs, repo.lastSyncFolderIDs)
}

// --- MoveImageFolder ---

func TestImageUsecase_MoveImageFolder_NoOp(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	item, err := uc.MoveImageFolder(context.Background(), imageID, "kp_abc123", &folderID, &folderID)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, 0, repo.moveFolderCalls)
}

func TestImageUsecase_MoveImageFolder_Moves(t *testing.T) {
	imageID := uuid.New()
	from, to := uuid.New(), uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, &mockStorageService{}, &mockThumbnailService{}, nil, nil, nil)

	item, err := uc.MoveImageFolder(context.Background(), imageID, "kp_abc123", &from, &to)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, 1, repo.moveFolderCalls)
	assert.Equal(t, imageID, repo.lastMoveImageID)
	assert.Equal(t, from, *repo.lastMoveFromFolderID)
	assert.Equal(t, to, *repo.lastMoveToFolderID)
}

// --- CleanupStaleUploads ---

func TestImageUsecase_CleanupStaleUploads_DeletesAll(t *testing.T) {
	stale := []*domain.PendingUpload{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
		{ID: uuid.New(), UserID: "kp_u2", R2Path: "users/kp_u2/images/b.jpg"},
	}
	pendingRepo := &mockPendingUploadRepository{pendings: stale}
	store := &mockStorageService{}
	uc := newImageUsecase(&mockImageRepository{}, pendingRepo, nil, store, &mockThumbnailService{}, nil, nil, nil)

	err := uc.CleanupStaleUploads(context.Background(), 30*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_u1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, "users/kp_u2/images/b.jpg")
	assert.Equal(t, 2, pendingRepo.deleteCalls)
}

func TestImageUsecase_CleanupStaleUploads_StorageErrorContinues(t *testing.T) {
	stale := []*domain.PendingUpload{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
	}
	pendingRepo := &mockPendingUploadRepository{pendings: stale}
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newImageUsecase(&mockImageRepository{}, pendingRepo, nil, store, &mockThumbnailService{}, nil, nil, nil)

	err := uc.CleanupStaleUploads(context.Background(), 30*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 1, pendingRepo.deleteCalls)
}

// --- PurgeExpiredTrash ---

func TestImageUsecase_PurgeExpiredTrash_WithThumbnail(t *testing.T) {
	thumbnailPath := "users/kp_u1/thumbnails/b.jpg"
	expired := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg", ThumbnailPath: &thumbnailPath},
	}
	repo := &mockImageRepository{images: expired}
	store := &mockStorageService{}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.NoError(t, err)
	assert.Equal(t, 2, store.deleteCalls)
	assert.Contains(t, store.deletedKeys, "users/kp_u1/images/a.jpg")
	assert.Contains(t, store.deletedKeys, thumbnailPath)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}

func TestImageUsecase_PurgeExpiredTrash_WithoutThumbnail(t *testing.T) {
	expired := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
	}
	repo := &mockImageRepository{images: expired}
	store := &mockStorageService{}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.NoError(t, err)
	assert.Equal(t, 1, store.deleteCalls)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}

func TestImageUsecase_PurgeExpiredTrash_StorageErrorContinues(t *testing.T) {
	expired := []*domain.Image{
		{ID: uuid.New(), UserID: "kp_u1", R2Path: "users/kp_u1/images/a.jpg"},
	}
	repo := &mockImageRepository{images: expired}
	store := &mockStorageService{deleteObjectErr: errors.New("r2 unavailable")}
	uc := newImageUsecase(repo, &mockPendingUploadRepository{}, nil, store, &mockThumbnailService{}, nil, nil, nil)

	err := uc.PurgeExpiredTrash(context.Background(), 30*24*time.Hour)

	require.NoError(t, err)
	assert.Equal(t, 1, repo.hardDeleteCalls)
}
