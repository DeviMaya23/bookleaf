package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type mockFolderRepository struct {
	folder        *domain.Folder
	folders       []*domain.Folder
	err           error
	deleteErr     error
	createdFolder *domain.Folder
	updatedFolder *domain.Folder
	listedUserID  string
	getByIDID     uuid.UUID
	getByIDUserID string
	deletedID     uuid.UUID
	deletedUserID string
}

func (m *mockFolderRepository) Create(_ context.Context, folder *domain.Folder) (*domain.Folder, error) {
	m.createdFolder = folder
	return m.folder, m.err
}

func (m *mockFolderRepository) List(_ context.Context, userID string) ([]*domain.Folder, error) {
	m.listedUserID = userID
	return m.folders, m.err
}

func (m *mockFolderRepository) GetByID(_ context.Context, id uuid.UUID, userID string) (*domain.Folder, error) {
	m.getByIDID = id
	m.getByIDUserID = userID
	return m.folder, m.err
}

func (m *mockFolderRepository) Update(_ context.Context, folder *domain.Folder) (*domain.Folder, error) {
	m.updatedFolder = folder
	return m.folder, m.err
}

func (m *mockFolderRepository) DeleteWithCascade(_ context.Context, id uuid.UUID, userID string) error {
	m.deletedID = id
	m.deletedUserID = userID
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return m.err
}

func TestFolderUsecase_Create_HappyPath(t *testing.T) {
	parentID := uuid.New()
	repo := &mockFolderRepository{
		folder: &domain.Folder{
			ID:       uuid.New(),
			UserID:   "kp_abc123",
			Name:     "travel",
			ParentID: &parentID,
		},
	}
	uc := NewFolderUsecase(repo)

	folder, err := uc.Create(context.Background(), "kp_abc123", "travel", &parentID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder.Name != "travel" {
		t.Fatalf("expected folder name travel, got %s", folder.Name)
	}
	if repo.createdFolder == nil {
		t.Fatal("expected repository Create to be called")
	}
	if repo.createdFolder.UserID != "kp_abc123" {
		t.Fatalf("expected user id kp_abc123, got %s", repo.createdFolder.UserID)
	}
}

func TestFolderUsecase_Create_ErrorPath(t *testing.T) {
	repo := &mockFolderRepository{}
	uc := NewFolderUsecase(repo)

	_, err := uc.Create(context.Background(), "kp_abc123", "   ", nil)
	if !errors.Is(err, ErrInvalidFolderName) {
		t.Fatalf("expected ErrInvalidFolderName, got: %v", err)
	}
}

func TestFolderUsecase_List_HappyPath(t *testing.T) {
	repo := &mockFolderRepository{
		folders: []*domain.Folder{
			{ID: uuid.New(), UserID: "kp_abc123", Name: "travel"},
		},
	}
	uc := NewFolderUsecase(repo)

	folders, err := uc.List(context.Background(), "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(folders))
	}
	if repo.listedUserID != "kp_abc123" {
		t.Fatalf("expected user id kp_abc123, got %s", repo.listedUserID)
	}
}

func TestFolderUsecase_List_ErrorPath(t *testing.T) {
	repo := &mockFolderRepository{
		err: errors.New("db error"),
	}
	uc := NewFolderUsecase(repo)

	_, err := uc.List(context.Background(), "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFolderUsecase_GetByID_HappyPath(t *testing.T) {
	folderID := uuid.New()
	repo := &mockFolderRepository{
		folder: &domain.Folder{ID: folderID, UserID: "kp_abc123", Name: "travel"},
	}
	uc := NewFolderUsecase(repo)

	folder, err := uc.GetByID(context.Background(), folderID, "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder.ID != folderID {
		t.Fatalf("expected folder id %s, got %s", folderID, folder.ID)
	}
	if repo.getByIDID != folderID {
		t.Fatalf("expected repo called with id %s, got %s", folderID, repo.getByIDID)
	}
}

func TestFolderUsecase_GetByID_ErrorPath(t *testing.T) {
	folderID := uuid.New()
	repo := &mockFolderRepository{
		err: errors.New("db error"),
	}
	uc := NewFolderUsecase(repo)

	_, err := uc.GetByID(context.Background(), folderID, "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFolderUsecase_Update_HappyPath(t *testing.T) {
	folderID := uuid.New()
	parentID := uuid.New()
	repo := &mockFolderRepository{
		folder: &domain.Folder{
			ID:       folderID,
			UserID:   "kp_abc123",
			Name:     "updated",
			ParentID: &parentID,
		},
	}
	uc := NewFolderUsecase(repo)

	folder, err := uc.Update(context.Background(), folderID, "kp_abc123", "updated", &parentID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if folder.Name != "updated" {
		t.Fatalf("expected name updated, got %s", folder.Name)
	}
	if repo.updatedFolder == nil {
		t.Fatal("expected repository Update to be called")
	}
	if repo.updatedFolder.ID != folderID {
		t.Fatalf("expected update id %s, got %s", folderID, repo.updatedFolder.ID)
	}
}

func TestFolderUsecase_Update_ErrorPath(t *testing.T) {
	folderID := uuid.New()
	repo := &mockFolderRepository{}
	uc := NewFolderUsecase(repo)

	_, err := uc.Update(context.Background(), folderID, "kp_abc123", "", nil)
	if !errors.Is(err, ErrInvalidFolderName) {
		t.Fatalf("expected ErrInvalidFolderName, got: %v", err)
	}
}

func TestFolderUsecase_Delete_HappyPath(t *testing.T) {
	folderID := uuid.New()
	repo := &mockFolderRepository{}
	uc := NewFolderUsecase(repo)

	err := uc.Delete(context.Background(), folderID, "kp_abc123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deletedID != folderID {
		t.Fatalf("expected delete id %s, got %s", folderID, repo.deletedID)
	}
	if repo.deletedUserID != "kp_abc123" {
		t.Fatalf("expected delete user kp_abc123, got %s", repo.deletedUserID)
	}
}

func TestFolderUsecase_Delete_ErrorPath(t *testing.T) {
	folderID := uuid.New()
	repo := &mockFolderRepository{
		deleteErr: errors.New("db error"),
	}
	uc := NewFolderUsecase(repo)

	err := uc.Delete(context.Background(), folderID, "kp_abc123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
