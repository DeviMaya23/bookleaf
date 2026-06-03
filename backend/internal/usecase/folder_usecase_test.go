package usecase

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- spies ---

type stubFolderRepo struct {
	folder *domain.Folder
	err    error
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
func (s *stubFolderRepo) Update(_ context.Context, _ *domain.Folder) (*domain.Folder, error) {
	return s.folder, s.err
}
func (s *stubFolderRepo) CountImagesByFolder(_ context.Context, _ uuid.UUID, _ string) (int, error) {
	return 0, s.err
}
func (s *stubFolderRepo) DeleteWithCascade(_ context.Context, _ uuid.UUID, _ string) error {
	return s.err
}

type stubImageCounter struct {
	count int64
	err   error
}

func (s *stubImageCounter) CountByFolderID(_ context.Context, _ uuid.UUID) (int64, error) {
	return s.count, s.err
}

func newTestFolderUsecase(folderRepo FolderRepository, counter ImageCounter) *folderUsecase {
	return NewFolderUsecase(folderRepo, counter, observability.NewTelemetry(nil, nil, nil))
}

// --- tests ---

func TestFolderUsecase_Create_BlankName(t *testing.T) {
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubImageCounter{})

	_, err := uc.Create(context.Background(), "kp_abc123", "   ", nil, nil)

	require.ErrorIs(t, err, ErrInvalidFolderName)
}

func TestFolderUsecase_Update_BlankName(t *testing.T) {
	uc := newTestFolderUsecase(&stubFolderRepo{}, &stubImageCounter{})

	_, err := uc.Update(context.Background(), uuid.New(), "kp_abc123", "", nil, nil)

	require.ErrorIs(t, err, ErrInvalidFolderName)
}

func TestFolderUsecase_GetByID(t *testing.T) {
	folderID := uuid.New()
	uc := newTestFolderUsecase(
		&stubFolderRepo{folder: &domain.Folder{ID: folderID, Name: "travel"}},
		&stubImageCounter{count: 3},
	)

	detail, err := uc.GetByID(context.Background(), folderID, "kp_abc123")

	require.NoError(t, err)
	assert.Equal(t, folderID, detail.Folder.ID)
	assert.EqualValues(t, 3, detail.ImageCount)
}
