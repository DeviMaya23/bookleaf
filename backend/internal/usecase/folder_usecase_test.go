package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- spies ---

type stubFolderRepo struct {
	folder       *domain.Folder
	err          error
	updateFields map[string]any
}

func (s *stubFolderRepo) Create(_ context.Context, _ *domain.Folder) (*domain.Folder, error) {
	return s.folder, s.err
}
func (s *stubFolderRepo) List(_ context.Context, _ string) ([]*domain.Folder, error) {
	return nil, s.err
}
func (s *stubFolderRepo) FindByName(_ context.Context, _, _ string) (*domain.Folder, error) {
	return s.folder, s.err
}
func (s *stubFolderRepo) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Folder, error) {
	return s.folder, s.err
}
func (s *stubFolderRepo) Update(_ context.Context, _ uuid.UUID, _ string, fields map[string]any) (*domain.Folder, error) {
	s.updateFields = fields
	return s.folder, s.err
}
func (s *stubFolderRepo) CountImagesByFolder(_ context.Context, _ uuid.UUID, _ string) (int, error) {
	return 0, s.err
}
func (s *stubFolderRepo) DeleteWithCascade(_ context.Context, _ uuid.UUID, _ string) error {
	return s.err
}
func (s *stubFolderRepo) ClearAllParents(_ context.Context, _ string) error {
	return s.err
}
func (s *stubFolderRepo) DeleteAllByUserID(_ context.Context, _ string) error {
	return s.err
}

type stubFolderImageRepo struct {
	count  int64
	images []*domain.Image
	err    error
}

func (s *stubFolderImageRepo) CountByFolderID(_ context.Context, _ uuid.UUID) (int64, error) {
	return s.count, s.err
}

func (s *stubFolderImageRepo) ListByFolder(_ context.Context, _ string, _ uuid.UUID, _ *string, _ *string) ([]*domain.Image, error) {
	return s.images, s.err
}

type stubStorageService struct {
	StorageService
}

// spyExportStorage is a value-return spy for StorageService.GetObject, keyed
// by R2 object path.
type spyExportStorage struct {
	StorageService
	objects map[string][]byte
	errs    map[string]error
}

func (s *spyExportStorage) GetObject(_ context.Context, key string) (io.ReadCloser, error) {
	if err, ok := s.errs[key]; ok {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(s.objects[key])), nil
}

func newTestFolderUsecase(folderRepo FolderRepository, imageRepo FolderImageRepository) *folderUsecase {
	return NewFolderUsecase(folderRepo, imageRepo, &stubStorageService{}, observability.NewTelemetry(nil, nil, nil))
}

// --- tests ---

func TestFolderUsecase_Create_BlankName(t *testing.T) {
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{})

	_, err := uc.Create(context.Background(), "kp_abc123", "   ", nil, nil)

	require.ErrorIs(t, err, ErrInvalidFolderName)
}

func TestFolderUsecase_Update_BlankName(t *testing.T) {
	repo := &stubFolderRepo{}
	uc := newTestFolderUsecase(repo, &stubFolderImageRepo{})
	blank := "   "

	_, err := uc.Update(context.Background(), uuid.New(), "kp_abc123", UpdateFolderParams{Name: &blank})

	require.ErrorIs(t, err, ErrInvalidFolderName)
	assert.Nil(t, repo.updateFields)
}

func TestFolderUsecase_Update_PassesThroughOnlyProvidedFields(t *testing.T) {
	folderID := uuid.New()
	name := "updated"
	repo := &stubFolderRepo{folder: &domain.Folder{ID: folderID, Name: name}}
	uc := newTestFolderUsecase(repo, &stubFolderImageRepo{})

	_, err := uc.Update(context.Background(), folderID, "kp_abc123", UpdateFolderParams{Name: &name})

	require.NoError(t, err)
	assert.Equal(t, map[string]any{"name": name}, repo.updateFields)
}

func TestFolderUsecase_Update_NotFound(t *testing.T) {
	name := "updated"
	repo := &stubFolderRepo{err: fmt.Errorf("update folder: %w", gorm.ErrRecordNotFound)}
	uc := newTestFolderUsecase(repo, &stubFolderImageRepo{})

	_, err := uc.Update(context.Background(), uuid.New(), "kp_abc123", UpdateFolderParams{Name: &name})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestFolderUsecase_GetByID(t *testing.T) {
	folderID := uuid.New()
	uc := newTestFolderUsecase(
		&stubFolderRepo{folder: &domain.Folder{ID: folderID, Name: "travel"}},
		&stubFolderImageRepo{count: 3},
	)

	detail, err := uc.GetByID(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, folderID, detail.Folder.ID)
	assert.EqualValues(t, 3, detail.ImageCount)
}

// --- ExportFolder ---

func zipEntryNames(t *testing.T, data []byte) []string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}

func TestFolderUsecase_ExportFolder_WritesEntriesWithDerivedNames(t *testing.T) {
	images := []*domain.Image{
		{ID: uuid.New(), Title: "Sunset", MIMEType: "image/jpeg", R2Path: "path/sunset"},
		{ID: uuid.New(), Title: "Portrait", MIMEType: "image/png", R2Path: "path/portrait"},
	}
	store := &spyExportStorage{objects: map[string][]byte{
		"path/sunset":   []byte("sunset-bytes"),
		"path/portrait": []byte("portrait-bytes"),
	}}
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{images: images})
	uc.store = store

	var buf bytes.Buffer
	err := uc.ExportFolder(context.Background(), uuid.New(), "kp_abc123", &buf)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Sunset.jpg", "Portrait.png"}, zipEntryNames(t, buf.Bytes()))
}

func TestFolderUsecase_ExportFolder_DeduplicatesCollidingNames(t *testing.T) {
	images := []*domain.Image{
		{ID: uuid.New(), Title: "Untitled", MIMEType: "image/jpeg", R2Path: "path/1"},
		{ID: uuid.New(), Title: "Untitled", MIMEType: "image/jpeg", R2Path: "path/2"},
	}
	store := &spyExportStorage{objects: map[string][]byte{
		"path/1": []byte("a"),
		"path/2": []byte("b"),
	}}
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{images: images})
	uc.store = store

	var buf bytes.Buffer
	err := uc.ExportFolder(context.Background(), uuid.New(), "kp_abc123", &buf)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Untitled.jpg", "Untitled (1).jpg"}, zipEntryNames(t, buf.Bytes()))
}

func TestFolderUsecase_ExportFolder_SanitizesTitleWithPathSeparators(t *testing.T) {
	images := []*domain.Image{
		{ID: uuid.New(), Title: "Trip/Day 1", MIMEType: "image/jpeg", R2Path: "path/1"},
	}
	store := &spyExportStorage{objects: map[string][]byte{"path/1": []byte("a")}}
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{images: images})
	uc.store = store

	var buf bytes.Buffer
	err := uc.ExportFolder(context.Background(), uuid.New(), "kp_abc123", &buf)

	require.NoError(t, err)
	names := zipEntryNames(t, buf.Bytes())
	require.Len(t, names, 1)
	assert.NotContains(t, names[0], "/")
	assert.NotContains(t, names[0], "\\")
}

func TestFolderUsecase_ExportFolder_EmptyFolderProducesValidEmptyZip(t *testing.T) {
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{images: nil})

	var buf bytes.Buffer
	err := uc.ExportFolder(context.Background(), uuid.New(), "kp_abc123", &buf)

	require.NoError(t, err)
	assert.Empty(t, zipEntryNames(t, buf.Bytes()))
}

func TestFolderUsecase_ExportFolder_ListByFolderError(t *testing.T) {
	listErr := errors.New("db error")
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{err: listErr})

	var buf bytes.Buffer
	err := uc.ExportFolder(context.Background(), uuid.New(), "kp_abc123", &buf)

	require.ErrorIs(t, err, listErr)
	assert.Zero(t, buf.Len())
}

func TestFolderUsecase_ExportFolder_GetObjectError(t *testing.T) {
	images := []*domain.Image{
		{ID: uuid.New(), Title: "Sunset", MIMEType: "image/jpeg", R2Path: "path/sunset"},
	}
	getErr := errors.New("r2 unavailable")
	store := &spyExportStorage{errs: map[string]error{"path/sunset": getErr}}
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubFolderImageRepo{images: images})
	uc.store = store

	var buf bytes.Buffer
	err := uc.ExportFolder(context.Background(), uuid.New(), "kp_abc123", &buf)

	require.ErrorIs(t, err, getErr)
}
