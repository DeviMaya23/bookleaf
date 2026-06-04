package usecase

import (
	"context"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// fakeFolderRepo is an in-memory ImageFolderRepository.
// Use it when the test needs to set up pre-state and assert post-state
// rather than assert which methods were called.
type fakeFolderRepo struct {
	folders map[string]*domain.Folder // keyed by lower(name)
	byID    map[uuid.UUID]*domain.Folder
}

func newFakeFolderRepo(initial ...*domain.Folder) *fakeFolderRepo {
	f := &fakeFolderRepo{
		folders: make(map[string]*domain.Folder),
		byID:    make(map[uuid.UUID]*domain.Folder),
	}
	for _, folder := range initial {
		f.add(folder)
	}
	return f
}

func (f *fakeFolderRepo) add(folder *domain.Folder) {
	f.folders[strings.ToLower(folder.Name)] = folder
	f.byID[folder.ID] = folder
}

func (f *fakeFolderRepo) GetByID(_ context.Context, id uuid.UUID, _ string) (*domain.Folder, error) {
	folder, ok := f.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return folder, nil
}

func (f *fakeFolderRepo) FindByName(_ context.Context, _, name string) (*domain.Folder, error) {
	folder, ok := f.folders[strings.ToLower(name)]
	if !ok {
		return nil, nil
	}
	return folder, nil
}

func (f *fakeFolderRepo) Create(_ context.Context, folder *domain.Folder) (*domain.Folder, error) {
	if folder.ID == uuid.Nil {
		folder.ID = uuid.New()
	}
	f.add(folder)
	return folder, nil
}
