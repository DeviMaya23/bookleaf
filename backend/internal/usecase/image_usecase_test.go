package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- test doubles ---

type mockImageRepository struct {
	image                *domain.Image
	images               []*domain.Image
	err                  error
	createdImage         *domain.Image
	updateFields         map[string]any
	lastUpdateID         uuid.UUID
	lastUpdateBy         string
	setFolderCalls       int
	setFolderImageID     uuid.UUID
	setFolderFolderID    *uuid.UUID
	syncFolderCalls      int
	lastSyncImageID      uuid.UUID
	lastSyncFolderIDs    []uuid.UUID
	moveFolderCalls      int
	lastMoveImageID      uuid.UUID
	lastMoveFromFolderID *uuid.UUID
	lastMoveToFolderID   *uuid.UUID
	updateAILabelsCalls  int
	lastAILabels         json.RawMessage
	lastListName         *string
	lastListSort         *string
	lastListDirection    *string
	lastListUnfiled      bool
	lastListFolderIDs    []uuid.UUID
	lastListTagIDs       []uuid.UUID
	lastListMIMETypes    []string
	listByFolderCalls    int
	lastListByFolderID   uuid.UUID
	listByFolderImages   []*domain.Image
}

func (m *mockImageRepository) Create(_ context.Context, img *domain.Image) (*domain.Image, error) {
	m.createdImage = img
	return m.image, m.err
}
func (m *mockImageRepository) List(_ context.Context, _ string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name *string, sortField *string, direction *string, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	m.lastListUnfiled = unfiled
	m.lastListFolderIDs = folderIDs
	m.lastListTagIDs = tagIDs
	m.lastListMIMETypes = mimeTypes
	m.lastListName = name
	m.lastListSort = sortField
	m.lastListDirection = direction
	return m.images, m.err
}
func (m *mockImageRepository) ListByFolder(_ context.Context, _ string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error) {
	m.listByFolderCalls++
	m.lastListByFolderID = folderID
	m.lastListSort = sortField
	m.lastListDirection = direction
	return m.listByFolderImages, m.err
}
func (m *mockImageRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}
func (m *mockImageRepository) UpdateThumbnailPath(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockImageRepository) UpdateAILabels(_ context.Context, _ uuid.UUID, labels json.RawMessage) error {
	m.updateAILabelsCalls++
	m.lastAILabels = labels
	return m.err
}
func (m *mockImageRepository) Update(_ context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error) {
	m.lastUpdateID = id
	m.lastUpdateBy = userID
	m.updateFields = _mapCopy(fields)
	return m.image, m.err
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

type mockStorageService struct {
	putURL          string
	getURL          string
	downloadURL     string
	err             error
	getObjectErr    error
	putObjectErr    error
	deleteObjectErr error
	objectBytes     []byte
	getCalls        int
	putCalls        int
	deleteCalls     int
	deletedKeys     []string

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

func noopTel() *observability.Telemetry {
	return observability.NewTelemetry(nil, nil, nil)
}

func newImageUsecase(imageRepo ImageRepository, tagRepo TagRepository, store StorageService) *imageUsecase {
	return NewImageUsecase(imageRepo, tagRepo, &stubImageFolderRepo{folder: &domain.Folder{ID: uuid.New()}}, store, noopTel())
}

func newImageUsecaseWithFolderRepo(imageRepo ImageRepository, tagRepo TagRepository, store StorageService, folderRepo ImageFolderRepository) *imageUsecase {
	return NewImageUsecase(imageRepo, tagRepo, folderRepo, store, noopTel())
}

// --- ListImages ---

func TestImageUsecase_ListImages_PassesFolderIDsTagIDsMIMETypesToRepository(t *testing.T) {
	folderIDs := []uuid.UUID{uuid.New(), uuid.New()}
	tagIDs := []uuid.UUID{uuid.New(), uuid.New()}
	mimeTypes := []string{"image/jpeg", "image/png"}
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	_, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{FolderIDs: folderIDs, TagIDs: tagIDs, MIMETypes: mimeTypes})

	require.NoError(t, err)
	assert.Equal(t, folderIDs, repo.lastListFolderIDs)
	assert.Equal(t, tagIDs, repo.lastListTagIDs)
	assert.Equal(t, mimeTypes, repo.lastListMIMETypes)
}

func TestImageUsecase_ListImages_NextCursor(t *testing.T) {
	now := time.Now().UTC()
	images := make([]*domain.Image, 11)
	for i := range images {
		images[i] = &domain.Image{ID: uuid.New(), CreatedAt: now.Add(-time.Duration(i) * time.Second)}
	}
	repo := &mockImageRepository{images: images}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, result.Images, 10)
	require.NotNil(t, result.NextCursor)
	assert.Equal(t, result.Images[9].Image.ID, result.NextCursor.ID)
}

func TestImageUsecase_ListImages_LastPage(t *testing.T) {
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}, {ID: uuid.New()}}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Limit: 10})

	require.NoError(t, err)
	assert.Len(t, result.Images, 2)
	assert.Nil(t, result.NextCursor)
}

func TestImageUsecase_ListImages_ThumbnailURL(t *testing.T) {
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New(), ThumbnailPath: &thumbnailPath}}}
	store := &mockStorageService{getURL: "https://r2.example.com/thumb"}
	uc := newImageUsecase(repo, nil, store)

	result, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{})

	require.NoError(t, err)
	require.Len(t, result.Images, 1)
	require.NotNil(t, result.Images[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/thumb", *result.Images[0].ThumbnailURL)
}

func TestImageUsecase_ListImages_PassesNameToRepository(t *testing.T) {
	repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})
	name := "heartopia"

	_, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{Name: &name})

	require.NoError(t, err)
	require.NotNil(t, repo.lastListName)
	assert.Equal(t, "heartopia", *repo.lastListName)
}

func TestImageUsecase_ListImages_SkipsBlankName(t *testing.T) {
	empty := ""
	tests := []struct {
		name   string
		params ListImagesParams
	}{
		{name: "nil name", params: ListImagesParams{}},
		{name: "empty string name", params: ListImagesParams{Name: &empty}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockImageRepository{images: []*domain.Image{{ID: uuid.New()}}}
			uc := newImageUsecase(repo, nil, &mockStorageService{})

			_, err := uc.ListImages(context.Background(), "kp_abc123", tt.params)

			require.NoError(t, err)
			assert.Nil(t, repo.lastListName)
		})
	}
}

// --- ListFolderImages ---

func TestImageUsecase_ListFolderImages_FolderFoundAndOwned(t *testing.T) {
	folderID := uuid.New()
	imageID := uuid.New()
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	position := "abc"
	repo := &mockImageRepository{
		listByFolderImages: []*domain.Image{
			{ID: imageID, ThumbnailPath: &thumbnailPath, ImageFolders: []domain.ImageFolder{{FolderID: folderID, Position: position}}},
		},
	}
	store := &mockStorageService{getURL: "https://r2.example.com/thumb"}
	folderRepo := &stubImageFolderRepo{folder: &domain.Folder{ID: folderID, UserID: "kp_abc123"}}
	uc := newImageUsecaseWithFolderRepo(repo, nil, store, folderRepo)

	items, err := uc.ListFolderImages(context.Background(), "kp_abc123", folderID, nil, nil)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.NotNil(t, items[0].FolderPosition)
	assert.Equal(t, position, *items[0].FolderPosition)
	require.NotNil(t, items[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/thumb", *items[0].ThumbnailURL)
}

func TestImageUsecase_ListFolderImages_FolderNotFoundOrNotOwned(t *testing.T) {
	folderID := uuid.New()
	repo := &mockImageRepository{listByFolderImages: []*domain.Image{{ID: uuid.New()}}}
	folderRepo := &stubImageFolderRepo{err: gorm.ErrRecordNotFound}
	uc := newImageUsecaseWithFolderRepo(repo, nil, &mockStorageService{}, folderRepo)

	_, err := uc.ListFolderImages(context.Background(), "kp_abc123", folderID, nil, nil)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Equal(t, 0, repo.listByFolderCalls)
}

// --- GetImage ---

func TestImageUsecase_GetImage_WithThumbnail(t *testing.T) {
	imageID := uuid.New()
	thumbnailPath := "users/kp_abc123/thumbnails/img.jpg"
	repo := &mockImageRepository{image: &domain.Image{ID: imageID, ThumbnailPath: &thumbnailPath}}
	store := &mockStorageService{getURL: "https://r2.example.com/view"}
	uc := newImageUsecase(repo, nil, store)

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
	uc := newImageUsecase(repo, nil, store)

	detail, err := uc.GetImage(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "https://r2.example.com/view", detail.ImageURL)
	assert.Nil(t, detail.ThumbnailURL)
}

func TestImageUsecase_GetImage_SuggestedFolderName(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name     string
		aiLabels json.RawMessage
		wantName *string
	}{
		{
			name:     "null ai_labels returns nil",
			aiLabels: nil,
			wantName: nil,
		},
		{
			name:     "empty ai_labels returns nil",
			aiLabels: json.RawMessage(`[]`),
			wantName: nil,
		},
		{
			name:     "populated ai_labels returns top label description",
			aiLabels: json.RawMessage(`[{"Description":"Nature","Score":0.98},{"Description":"Outdoor","Score":0.75}]`),
			wantName: strPtr("Nature"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockImageRepository{image: &domain.Image{ID: imageID, AILabels: tt.aiLabels}}
			store := &mockStorageService{getURL: "https://r2.example.com/view"}
			uc := newImageUsecase(repo, nil, store)

			detail, err := uc.GetImage(context.Background(), imageID, "kp_abc123")

			require.NoError(t, err)
			if tt.wantName == nil {
				assert.Nil(t, detail.SuggestedFolderName)
			} else {
				require.NotNil(t, detail.SuggestedFolderName)
				assert.Equal(t, *tt.wantName, *detail.SuggestedFolderName)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

func TestImageUsecase_GetImage_PresignFails(t *testing.T) {
	repo := &mockImageRepository{image: &domain.Image{ID: uuid.New()}}
	store := &mockStorageService{err: errors.New("presign failed")}
	uc := newImageUsecase(repo, nil, store)

	_, err := uc.GetImage(context.Background(), uuid.New(), "kp_abc123")

	require.Error(t, err)
}

// --- DownloadImage ---

func TestImageUsecase_DownloadImage(t *testing.T) {
	imageID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID, Title: "sunset", MIMEType: "image/jpeg", R2Path: "users/kp_abc123/images/photo.jpg"}}
	store := &mockStorageService{downloadURL: "https://r2.example.com/download"}
	uc := newImageUsecase(repo, nil, store)

	url, err := uc.DownloadImage(context.Background(), imageID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, "https://r2.example.com/download", url)
	assert.Equal(t, "sunset.jpg", store.lastDownloadFilename)
	assert.Equal(t, downloadURLTTL, store.lastDownloadTTL)
}

// --- UpdateImage ---

func TestImageUsecase_UpdateImage_FieldsAssembled(t *testing.T) {
	imageID := uuid.New()
	title := "new title"
	description := "new description"
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

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
	uc := newImageUsecase(repo, tagRepo, &mockStorageService{})

	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{Tags: nil})

	require.NoError(t, err)
	assert.Equal(t, 0, tagRepo.replaceCalls)
}

func TestImageUsecase_UpdateImage_WithTags(t *testing.T) {
	imageID := uuid.New()
	tag1, tag2 := uuid.New(), uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	tagRepo := &mockTagRepository{}
	uc := newImageUsecase(repo, tagRepo, &mockStorageService{})

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
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	_, err := uc.UpdateImage(context.Background(), imageID, "kp_abc123", UpdateImageParams{FolderIDs: nil})

	require.NoError(t, err)
	assert.Equal(t, 0, repo.syncFolderCalls)
}

func TestImageUsecase_UpdateImage_WithFolderIDs(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

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
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	item, err := uc.MoveImageFolder(context.Background(), imageID, "kp_abc123", &folderID, &folderID)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, 0, repo.moveFolderCalls)
}

func TestImageUsecase_MoveImageFolder_Moves(t *testing.T) {
	imageID := uuid.New()
	from, to := uuid.New(), uuid.New()
	repo := &mockImageRepository{image: &domain.Image{ID: imageID}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	item, err := uc.MoveImageFolder(context.Background(), imageID, "kp_abc123", &from, &to)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, 1, repo.moveFolderCalls)
	assert.Equal(t, imageID, repo.lastMoveImageID)
	assert.Equal(t, from, *repo.lastMoveFromFolderID)
	assert.Equal(t, to, *repo.lastMoveToFolderID)
}


func TestImageUsecase_ListImages_PassesSortAndDirection(t *testing.T) {
	sortVal := "title"
	dirVal := "asc"

	repo := &mockImageRepository{images: []*domain.Image{}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	_, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{
		Sort:      &sortVal,
		Direction: &dirVal,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.lastListSort)
	assert.Equal(t, "title", *repo.lastListSort)
	require.NotNil(t, repo.lastListDirection)
	assert.Equal(t, "asc", *repo.lastListDirection)
}

func TestImageUsecase_ListImages_PassesNilSortAndDirection(t *testing.T) {
	repo := &mockImageRepository{images: []*domain.Image{}}
	uc := newImageUsecase(repo, nil, &mockStorageService{})

	_, err := uc.ListImages(context.Background(), "kp_abc123", ListImagesParams{
		Sort:      nil,
		Direction: nil,
	})

	require.NoError(t, err)
	assert.Nil(t, repo.lastListSort)
	assert.Nil(t, repo.lastListDirection)
}
