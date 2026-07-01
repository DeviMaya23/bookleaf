package usecase

import (
	"context"
	"errors"
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

// fakeFolderShareRepo is an in-memory FolderShareRepository.
// CreateShare reads existing state (GetByFolderID) to decide whether to
// create a new row, so this is a fake rather than a spy.
type fakeFolderShareRepo struct {
	byFolderID map[uuid.UUID]*domain.FolderShare
	byToken    map[string]*domain.FolderShare

	// firstGetByFolderIDNotFound makes the first GetByFolderID call return
	// not-found regardless of pre-seeded state, simulating a row that does
	// not exist yet when CreateShare first checks.
	firstGetByFolderIDNotFound bool
	getByFolderIDCalls         int

	// failCreateWithUniqueViolation makes Create return a Postgres unique
	// constraint violation without inserting a row.
	failCreateWithUniqueViolation bool
	createCalls                   int
}

func newFakeFolderShareRepo(initial ...*domain.FolderShare) *fakeFolderShareRepo {
	f := &fakeFolderShareRepo{
		byFolderID: make(map[uuid.UUID]*domain.FolderShare),
		byToken:    make(map[string]*domain.FolderShare),
	}
	for _, share := range initial {
		f.add(share)
	}
	return f
}

func (f *fakeFolderShareRepo) add(share *domain.FolderShare) {
	f.byFolderID[share.FolderID] = share
	f.byToken[share.Token] = share
}

func (f *fakeFolderShareRepo) Create(_ context.Context, folderID uuid.UUID, token string) (*domain.FolderShare, error) {
	f.createCalls++
	if f.failCreateWithUniqueViolation {
		return nil, errors.New(`ERROR: duplicate key value violates unique constraint "folder_shares_folder_id_key" (SQLSTATE 23505)`)
	}
	share := &domain.FolderShare{ID: uuid.New(), FolderID: folderID, Token: token}
	f.add(share)
	return share, nil
}

func (f *fakeFolderShareRepo) GetByFolderID(_ context.Context, folderID uuid.UUID) (*domain.FolderShare, error) {
	f.getByFolderIDCalls++
	if f.firstGetByFolderIDNotFound && f.getByFolderIDCalls == 1 {
		return nil, gorm.ErrRecordNotFound
	}
	share, ok := f.byFolderID[folderID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return share, nil
}

func (f *fakeFolderShareRepo) GetByToken(_ context.Context, token string) (*domain.FolderShare, error) {
	share, ok := f.byToken[token]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return share, nil
}

func (f *fakeFolderShareRepo) DeleteByFolderID(_ context.Context, folderID uuid.UUID) error {
	if share, ok := f.byFolderID[folderID]; ok {
		delete(f.byFolderID, folderID)
		delete(f.byToken, share.Token)
	}
	return nil
}

// fakeCategorisationLogRepo is an in-memory categorisationLogRepository.
// GetByImageID returns pre-seeded state; Create captures the written entry.
// Use it when the test pre-seeds an existing log and then asserts what was
// (or was not) written afterwards.
type fakeCategorisationLogRepo struct {
	existing *domain.CategorisationLog
	created  *domain.CategorisationLog
}

func (f *fakeCategorisationLogRepo) GetByImageID(_ context.Context, _ uuid.UUID) (*domain.CategorisationLog, error) {
	return f.existing, nil
}

func (f *fakeCategorisationLogRepo) Create(_ context.Context, log *domain.CategorisationLog) error {
	f.created = log
	return nil
}
