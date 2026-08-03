package usecase

import (
	"context"
	"encoding/json"
	"errors"
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
func (m *mockPendingUploadRepository) ListAllByUserID(_ context.Context, _ string) ([]*domain.PendingUpload, error) {
	return m.pendings, m.err
}
func (m *mockPendingUploadRepository) DeleteAllByUserID(_ context.Context, _ string) error {
	return m.err
}

type stubUserRepo struct{ user *domain.User }

func (s *stubUserRepo) GetOrCreate(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) SetAccountState(_ context.Context, _ string, _ domain.AccountState) error {
	return nil
}
func (s *stubUserRepo) MarkPurged(_ context.Context, _ string, _ time.Time) error {
	return nil
}
func (s *stubUserRepo) ListByAccountState(_ context.Context, _ domain.AccountState) ([]*domain.User, error) {
	return nil, nil
}
func (s *stubUserRepo) ListPurgedBefore(_ context.Context, _ time.Time) ([]*domain.User, error) {
	return nil, nil
}
func (s *stubUserRepo) HardDelete(_ context.Context, _ string) error {
	return nil
}
func (s *stubUserRepo) Update(_ context.Context, _ string, _ map[string]any) (*domain.User, error) {
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

func intPtr(v int) *int {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
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

func (n *noopJobEnqueuer) Insert(_ context.Context, _ JobArgs) error       { return nil }
func (n *noopJobEnqueuer) InsertUnique(_ context.Context, _ JobArgs) error { return nil }

type capturingJobEnqueuer struct {
	inserted []JobArgs
}

func (c *capturingJobEnqueuer) Insert(_ context.Context, args JobArgs) error {
	c.inserted = append(c.inserted, args)
	return nil
}
func (c *capturingJobEnqueuer) InsertUnique(_ context.Context, args JobArgs) error {
	c.inserted = append(c.inserted, args)
	return nil
}

type spyJobEnqueuer struct {
	insertCalls int
}

func (s *spyJobEnqueuer) Insert(_ context.Context, _ JobArgs) error {
	s.insertCalls++
	return nil
}
func (s *spyJobEnqueuer) InsertUnique(_ context.Context, _ JobArgs) error {
	s.insertCalls++
	return nil
}

type failAfterJobEnqueuer struct {
	failOnCall int
	calls      int
	inserted   []JobArgs
}

func (f *failAfterJobEnqueuer) Insert(_ context.Context, args JobArgs) error {
	f.calls++
	if f.calls == f.failOnCall {
		return errors.New("enqueue failed")
	}
	f.inserted = append(f.inserted, args)
	return nil
}
func (f *failAfterJobEnqueuer) InsertUnique(_ context.Context, args JobArgs) error {
	return f.Insert(context.Background(), args)
}

func newImageUploadUsecase(
	imageRepo UploadImageRepository,
	pendingRepo PendingUploadRepository,
	folderRepo ImageFolderRepository,
	userRepo UserRepository,
	store StorageService,
	vision VisionService,
) *imageUploadUsecase {
	var hashRepo ImageHashRepository
	if mr, ok := imageRepo.(ImageHashRepository); ok {
		hashRepo = mr
	}
	return NewImageUploadUsecase(imageRepo, hashRepo, pendingRepo, folderRepo, userRepo, store, vision, &noopJobEnqueuer{}, noopTel())
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

func TestImageUploadUsecase_CompleteUpload_PersistsClientSuppliedValues(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/test.png", MIMEType: "image/png"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	store := &mockStorageService{}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), store, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", intPtr(1920), intPtr(1080), int64Ptr(245760), nil)

	require.NoError(t, err)
	require.NotNil(t, imageRepo.createdImage)
	require.NotNil(t, imageRepo.createdImage.Width)
	require.NotNil(t, imageRepo.createdImage.Height)
	require.NotNil(t, imageRepo.createdImage.FileSize)
	assert.Equal(t, 1920, *imageRepo.createdImage.Width)
	assert.Equal(t, 1080, *imageRepo.createdImage.Height)
	assert.EqualValues(t, 245760, *imageRepo.createdImage.FileSize)
	assert.Equal(t, 0, store.getCalls, "backend must not fetch image bytes from R2 to derive metadata")
}

func TestImageUploadUsecase_CompleteUpload_NonPositiveValuesStoredAsNull(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.bin", MIMEType: "application/octet-stream"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", intPtr(-1), intPtr(0), int64Ptr(1024), nil)

	require.NoError(t, err)
	assert.Nil(t, imageRepo.createdImage.Width)
	assert.Nil(t, imageRepo.createdImage.Height)
	require.NotNil(t, imageRepo.createdImage.FileSize)
	assert.EqualValues(t, 1024, *imageRepo.createdImage.FileSize)
}

func TestImageUploadUsecase_CompleteUpload_OmittedValuesStoredAsNull(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.bin", MIMEType: "application/octet-stream"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Nil(t, imageRepo.createdImage.Width)
	assert.Nil(t, imageRepo.createdImage.Height)
	assert.Nil(t, imageRepo.createdImage.FileSize)
}

func TestImageUploadUsecase_CompleteUpload_SetsFolderWhenPendingHasFolderID(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg", FolderID: &folderID},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, nil)

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
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, enqueuer, noopTel())

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, imageRepo.createdImage.ThumbnailPath)
	assert.True(t, strings.HasPrefix(*imageRepo.createdImage.ThumbnailPath, "users/kp_abc123/thumbnails/"))
	assert.Equal(t, 1, enqueuer.insertCalls, "only vision job should be enqueued")
}

func TestImageUploadUsecase_CompleteUpload_PhashProvided_DuplicatesFound(t *testing.T) {
	imageID := uuid.New()
	dupID := uuid.New()
	thumb := "users/kp_abc123/thumbnails/dup.jpg"
	phash := "0110110100110101010101010101010101010101010101010101010101010101"
	dup := &domain.Image{ID: dupID, Title: "duplicate image", ThumbnailPath: &thumb}
	imageRepo := &mockImageRepository{
		image:                &domain.Image{ID: imageID},
		findDuplicatesResult: []*domain.Image{dup},
	}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, &phash)

	require.NoError(t, err)
	require.Len(t, result.Duplicates, 1)
	assert.Equal(t, dupID, result.Duplicates[0].ID)
	assert.Equal(t, "duplicate image", result.Duplicates[0].Title)
	assert.Equal(t, &thumb, result.Duplicates[0].ThumbnailPath)
}

func TestImageUploadUsecase_CompleteUpload_PhashProvided_NoDuplicates(t *testing.T) {
	imageID := uuid.New()
	phash := "0110110100110101010101010101010101010101010101010101010101010101"
	imageRepo := &mockImageRepository{
		image:                &domain.Image{ID: imageID},
		findDuplicatesResult: []*domain.Image{},
	}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, &phash)

	require.NoError(t, err)
	assert.Empty(t, result.Duplicates)
	assert.Equal(t, 1, imageRepo.findDuplicatesCalls)
}

func TestImageUploadUsecase_CompleteUpload_PhashAbsent_FindDuplicatesNotCalled(t *testing.T) {
	imageID := uuid.New()
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	pendingRepo := &mockPendingUploadRepository{
		pending: &domain.PendingUpload{ID: imageID, UserID: "kp_abc123", R2Path: "test.jpg", MIMEType: "image/jpeg"},
	}
	pendingRepo.transactionImageRepo = imageRepo
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, noopTel())

	result, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, nil)

	require.NoError(t, err)
	assert.Empty(t, result.Duplicates)
	assert.Equal(t, 0, imageRepo.findDuplicatesCalls)
}

// --- ProcessVisionLabelling ---

func TestImageUploadUsecase_ProcessVisionLabelling_SavesLabels(t *testing.T) {
	imageID := uuid.New()
	thumb := "users/kp_abc123/thumbnails/thumb.jpg"
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/img.jpg", ThumbnailPath: &thumb}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{labels: []domain.Label{{Description: "Nature", Score: 0.98}}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc)

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 1, visionSvc.calls)
	assert.Equal(t, 1, imageRepo.updateLabelsCalls)
	var saved []domain.Label
	require.NoError(t, json.Unmarshal(imageRepo.lastAILabels, &saved))
	require.Len(t, saved, 1)
	assert.Equal(t, "Nature", saved[0].Description)
	require.Len(t, imageRepo.lastImageLabels, 1)
	assert.Equal(t, "Nature", imageRepo.lastImageLabels[0].Label)
	assert.Equal(t, float32(0.98), imageRepo.lastImageLabels[0].Score)
	assert.Equal(t, imageID, imageRepo.lastImageLabels[0].ImageID)
}

func TestImageUploadUsecase_ProcessVisionLabelling_ZeroLabelsWritesEmpty(t *testing.T) {
	imageID := uuid.New()
	thumb := "users/kp_abc123/thumbnails/thumb.jpg"
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/img.jpg", ThumbnailPath: &thumb}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{labels: []domain.Label{}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	uc := newImageUploadUsecase(imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc)

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 1, imageRepo.updateLabelsCalls)
	assert.Equal(t, "[]", string(imageRepo.lastAILabels))
	assert.Empty(t, imageRepo.lastImageLabels)
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

func TestImageUploadUsecase_ProcessVisionLabelling_AICategorisationEnabled_EnqueuesCategoriseJob(t *testing.T) {
	imageID := uuid.New()
	thumb := "users/kp_abc123/thumbnails/thumb.jpg"
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", ThumbnailPath: &thumb}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{labels: []domain.Label{{Description: "Nature", Score: 0.98}}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true, AICategorisationEnabled: true}}
	enqueuer := &capturingJobEnqueuer{}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc, enqueuer, noopTel())

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	require.Len(t, enqueuer.inserted, 1)
	args, ok := enqueuer.inserted[0].(CategoriseImageArgs)
	require.True(t, ok, "inserted job must be CategoriseImageArgs")
	assert.Equal(t, imageID, args.ImageID)
	assert.Equal(t, "kp_abc123", args.UserID)
}

func TestImageUploadUsecase_ProcessVisionLabelling_AICategorisationDisabled_NoCategoriseJob(t *testing.T) {
	imageID := uuid.New()
	thumb := "users/kp_abc123/thumbnails/thumb.jpg"
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", ThumbnailPath: &thumb}}
	store := &mockStorageService{objectBytes: []byte("img-bytes")}
	visionSvc := &mockVisionService{labels: []domain.Label{{Description: "Nature", Score: 0.98}}}
	userRepo := &stubUserRepo{user: &domain.User{ID: "kp_abc123", VisionEnabled: true, AICategorisationEnabled: false}}
	enqueuer := &capturingJobEnqueuer{}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, &mockPendingUploadRepository{}, nil, userRepo, store, visionSvc, enqueuer, noopTel())

	err := uc.ProcessVisionLabelling(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Empty(t, enqueuer.inserted)
}

func TestImageUploadUsecase_ProcessVisionLabelling_VisionAPIErrorReturnsError(t *testing.T) {
	imageID := uuid.New()
	thumb := "users/kp_abc123/thumbnails/thumb.jpg"
	imageRepo := &mockImageRepository{image: &domain.Image{ID: imageID, UserID: "kp_abc123", R2Path: "users/kp_abc123/images/img.jpg", ThumbnailPath: &thumb}}
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
	uc := NewImageUploadUsecase(imageRepo, imageRepo, pendingRepo, nil, defaultUserRepo(), &mockStorageService{}, nil, &noopJobEnqueuer{}, tel)

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, nil)
	require.NoError(t, err)

	points := findInt64Sum(collect(), "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)
	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "success", status.AsString())
}

func TestImageUploadUsecase_BackfillVisionLabels_AllEnqueued(t *testing.T) {
	images := []*domain.Image{{ID: uuid.New()}, {ID: uuid.New()}, {ID: uuid.New()}}
	imageRepo := &mockImageRepository{listUnlabelledResult: images}
	enqueuer := &spyJobEnqueuer{}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, &mockPendingUploadRepository{}, nil, defaultUserRepo(), &mockStorageService{}, nil, enqueuer, noopTel())

	count, err := uc.BackfillVisionLabels(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, 3, enqueuer.insertCalls)
}

func TestImageUploadUsecase_BackfillVisionLabels_NoUnlabelledImages(t *testing.T) {
	imageRepo := &mockImageRepository{listUnlabelledResult: nil}
	enqueuer := &spyJobEnqueuer{}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, &mockPendingUploadRepository{}, nil, defaultUserRepo(), &mockStorageService{}, nil, enqueuer, noopTel())

	count, err := uc.BackfillVisionLabels(context.Background(), "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, enqueuer.insertCalls)
}

func TestImageUploadUsecase_BackfillVisionLabels_ListErrorPropagated(t *testing.T) {
	imageRepo := &mockImageRepository{listUnlabelledErr: errors.New("db unavailable")}
	enqueuer := &spyJobEnqueuer{}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, &mockPendingUploadRepository{}, nil, defaultUserRepo(), &mockStorageService{}, nil, enqueuer, noopTel())

	count, err := uc.BackfillVisionLabels(context.Background(), "kp_abc123")

	require.Error(t, err)
	assert.ErrorContains(t, err, "list unlabelled images")
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, enqueuer.insertCalls)
}

func TestImageUploadUsecase_BackfillVisionLabels_EnqueueFailureStops(t *testing.T) {
	images := []*domain.Image{{ID: uuid.New()}, {ID: uuid.New()}, {ID: uuid.New()}}
	imageRepo := &mockImageRepository{listUnlabelledResult: images}
	enqueuer := &failAfterJobEnqueuer{failOnCall: 2}
	uc := NewImageUploadUsecase(imageRepo, imageRepo, &mockPendingUploadRepository{}, nil, defaultUserRepo(), &mockStorageService{}, nil, enqueuer, noopTel())

	count, err := uc.BackfillVisionLabels(context.Background(), "kp_abc123")

	require.Error(t, err)
	assert.ErrorContains(t, err, "enqueue vision labelling")
	assert.Equal(t, 1, count)
	assert.Equal(t, 2, enqueuer.calls)
	assert.Len(t, enqueuer.inserted, 1)
}

func TestImageUploadUsecase_CompleteUpload_UploadCount_Error(t *testing.T) {
	imageID := uuid.New()
	tel, collect := makeMetricsTel(t)
	pendingRepo := &mockPendingUploadRepository{getErr: gorm.ErrRecordNotFound}
	uc := NewImageUploadUsecase(&mockImageRepository{}, &mockImageRepository{}, pendingRepo, nil, nil, &mockStorageService{}, nil, &noopJobEnqueuer{}, tel)

	_, err := uc.CompleteUpload(context.Background(), imageID, "kp_abc123", nil, nil, nil, nil)
	require.Error(t, err)

	points := findInt64Sum(collect(), "r2.upload.count")
	require.Len(t, points, 1)
	assert.Equal(t, int64(1), points[0].Value)
	status, ok := points[0].Attributes.Value(attribute.Key("r2.status"))
	require.True(t, ok)
	assert.Equal(t, "error", status.AsString())
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
