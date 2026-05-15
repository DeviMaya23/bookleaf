package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockFolderRepository struct {
	folder  *domain.Folder
	folders []*domain.Folder
	count   int
	err     error
}

func (m *mockFolderRepository) Create(_ context.Context, _ *domain.Folder) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderRepository) List(_ context.Context, _ string) ([]*domain.Folder, error) {
	return m.folders, m.err
}

func (m *mockFolderRepository) FindByName(_ context.Context, _, _ string) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderRepository) Update(_ context.Context, _ *domain.Folder) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderRepository) CountImagesByFolder(_ context.Context, _ uuid.UUID, _ string) (int, error) {
	return m.count, m.err
}

func (m *mockFolderRepository) DeleteWithCascade(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

type mockFolderImageRepository struct {
	count int64
	err   error
}

func (m *mockFolderImageRepository) Create(_ context.Context, _ *domain.Image) (*domain.Image, error) {
	return nil, m.err
}
func (m *mockFolderImageRepository) List(_ context.Context, _ string, _ *uuid.UUID, _ bool, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	return nil, m.err
}
func (m *mockFolderImageRepository) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return nil, m.err
}
func (m *mockFolderImageRepository) GetDeletedByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return nil, m.err
}
func (m *mockFolderImageRepository) UpdateThumbnailPath(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockFolderImageRepository) UpdateAILabels(_ context.Context, _ uuid.UUID, _ json.RawMessage) error {
	return m.err
}
func (m *mockFolderImageRepository) Update(_ context.Context, _ uuid.UUID, _ string, _ map[string]any) (*domain.Image, error) {
	return nil, m.err
}
func (m *mockFolderImageRepository) SoftDelete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockFolderImageRepository) Restore(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}
func (m *mockFolderImageRepository) ListTrashed(_ context.Context, _ string, _ *ImageCursor, _ int) ([]*domain.Image, error) {
	return nil, m.err
}
func (m *mockFolderImageRepository) CountByFolderID(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.count, m.err
}

func newFolderUsecaseForTest(folderRepo *mockFolderRepository, imageRepo *mockFolderImageRepository) FolderUsecase {
	return NewFolderUsecase(folderRepo, imageRepo, observability.NewTelemetry(nil, nil, nil))
}

func TestFolderUsecase_Create(t *testing.T) {
	folderID := uuid.New()
	desc := "desc"

	t.Run("creates folder successfully", func(t *testing.T) {
		uc := newFolderUsecaseForTest(
			&mockFolderRepository{folder: &domain.Folder{ID: folderID, Name: "travel", Description: &desc}},
			&mockFolderImageRepository{},
		)

		folder, err := uc.Create(context.Background(), "kp_abc123", "travel", nil, &desc)

		require.NoError(t, err)
		assert.Equal(t, folderID, folder.ID)
		require.NotNil(t, folder.Description)
		assert.Equal(t, desc, *folder.Description)
	})

	t.Run("returns error for blank name", func(t *testing.T) {
		uc := newFolderUsecaseForTest(&mockFolderRepository{}, &mockFolderImageRepository{})

		_, err := uc.Create(context.Background(), "kp_abc123", "   ", nil, nil)

		require.ErrorIs(t, err, ErrInvalidFolderName)
	})
}

func TestFolderUsecase_List(t *testing.T) {
	t.Run("returns user folders", func(t *testing.T) {
		uc := newFolderUsecaseForTest(
			&mockFolderRepository{folders: []*domain.Folder{{ID: uuid.New()}, {ID: uuid.New()}}},
			&mockFolderImageRepository{},
		)

		folders, err := uc.List(context.Background(), "kp_abc123")

		require.NoError(t, err)
		assert.Len(t, folders, 2)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		uc := newFolderUsecaseForTest(&mockFolderRepository{err: errors.New("db error")}, &mockFolderImageRepository{})

		_, err := uc.List(context.Background(), "kp_abc123")

		require.Error(t, err)
	})
}

func TestFolderUsecase_GetByID(t *testing.T) {
	folderID := uuid.New()

	t.Run("returns folder detail with image count", func(t *testing.T) {
		uc := newFolderUsecaseForTest(
			&mockFolderRepository{folder: &domain.Folder{ID: folderID, Name: "travel"}},
			&mockFolderImageRepository{count: 3},
		)

		detail, err := uc.GetByID(context.Background(), folderID, "kp_abc123")

		require.NoError(t, err)
		assert.Equal(t, folderID, detail.Folder.ID)
		assert.EqualValues(t, 3, detail.ImageCount)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		uc := newFolderUsecaseForTest(&mockFolderRepository{err: errors.New("db error")}, &mockFolderImageRepository{})

		_, err := uc.GetByID(context.Background(), folderID, "kp_abc123")

		require.Error(t, err)
	})
}

func TestFolderUsecase_Update(t *testing.T) {
	folderID := uuid.New()
	desc := "new desc"

	t.Run("updates folder successfully", func(t *testing.T) {
		uc := newFolderUsecaseForTest(
			&mockFolderRepository{folder: &domain.Folder{ID: folderID, Name: "updated", Description: &desc}},
			&mockFolderImageRepository{},
		)

		folder, err := uc.Update(context.Background(), folderID, "kp_abc123", "updated", nil, &desc)

		require.NoError(t, err)
		assert.Equal(t, folderID, folder.ID)
		require.NotNil(t, folder.Description)
		assert.Equal(t, desc, *folder.Description)
	})

	t.Run("returns error for blank name", func(t *testing.T) {
		uc := newFolderUsecaseForTest(&mockFolderRepository{}, &mockFolderImageRepository{})

		_, err := uc.Update(context.Background(), folderID, "kp_abc123", "", nil, nil)

		require.ErrorIs(t, err, ErrInvalidFolderName)
	})
}

func TestFolderUsecase_Delete(t *testing.T) {
	folderID := uuid.New()

	t.Run("deletes folder successfully", func(t *testing.T) {
		uc := newFolderUsecaseForTest(&mockFolderRepository{}, &mockFolderImageRepository{})

		err := uc.Delete(context.Background(), folderID, "kp_abc123")

		require.NoError(t, err)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		uc := newFolderUsecaseForTest(&mockFolderRepository{err: errors.New("db error")}, &mockFolderImageRepository{})

		err := uc.Delete(context.Background(), folderID, "kp_abc123")

		require.Error(t, err)
	})
}
