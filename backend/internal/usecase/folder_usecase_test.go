package usecase

import (
	"context"
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

func TestFolderUsecase_Create(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name      string
		inputName string
		repo      *mockFolderRepository
		wantName  string
		wantErr   error
	}{
		{
			name:      "creates folder successfully",
			inputName: "travel",
			repo:      &mockFolderRepository{folder: &domain.Folder{ID: folderID, Name: "travel"}},
			wantName:  "travel",
		},
		{
			name:      "returns error for blank name",
			inputName: "   ",
			repo:      &mockFolderRepository{},
			wantErr:   ErrInvalidFolderName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFolderUsecase(tt.repo, observability.NewTelemetry(nil, nil, nil))

			folder, err := uc.Create(context.Background(), "kp_abc123", tt.inputName, nil)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, folder.Name)
		})
	}
}

func TestFolderUsecase_List(t *testing.T) {
	tests := []struct {
		name    string
		repo    *mockFolderRepository
		wantLen int
		wantErr bool
	}{
		{
			name: "returns user folders",
			repo: &mockFolderRepository{
				folders: []*domain.Folder{
					{ID: uuid.New(), Name: "travel"},
					{ID: uuid.New(), Name: "design"},
				},
			},
			wantLen: 2,
		},
		{
			name:    "propagates repository error",
			repo:    &mockFolderRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFolderUsecase(tt.repo, observability.NewTelemetry(nil, nil, nil))

			folders, err := uc.List(context.Background(), "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, folders, tt.wantLen)
		})
	}
}

func TestFolderUsecase_GetByID(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockFolderRepository
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name:   "returns folder by id",
			repo:   &mockFolderRepository{folder: &domain.Folder{ID: folderID, Name: "travel"}},
			wantID: folderID,
		},
		{
			name:    "propagates repository error",
			repo:    &mockFolderRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFolderUsecase(tt.repo, observability.NewTelemetry(nil, nil, nil))

			folder, err := uc.GetByID(context.Background(), folderID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, folder.ID)
		})
	}
}

func TestFolderUsecase_Update(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name      string
		inputName string
		repo      *mockFolderRepository
		wantName  string
		wantErr   error
	}{
		{
			name:      "updates folder successfully",
			inputName: "updated",
			repo:      &mockFolderRepository{folder: &domain.Folder{ID: folderID, Name: "updated"}},
			wantName:  "updated",
		},
		{
			name:      "returns error for blank name",
			inputName: "",
			repo:      &mockFolderRepository{},
			wantErr:   ErrInvalidFolderName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFolderUsecase(tt.repo, observability.NewTelemetry(nil, nil, nil))

			folder, err := uc.Update(context.Background(), folderID, "kp_abc123", tt.inputName, nil)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, folder.Name)
		})
	}
}

func TestFolderUsecase_Delete(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockFolderRepository
		wantErr bool
	}{
		{
			name: "deletes folder successfully",
			repo: &mockFolderRepository{},
		},
		{
			name:    "propagates repository error",
			repo:    &mockFolderRepository{err: errors.New("db error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFolderUsecase(tt.repo, observability.NewTelemetry(nil, nil, nil))

			err := uc.Delete(context.Background(), folderID, "kp_abc123")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
