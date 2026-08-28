package repository

import (
	"context"
	"testing"

	"github.com/devi/bookleaf/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- Create / GetByFolderID ---

func TestFolderShareRepository_Create_PersistsAndGetByFolderID(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	repo := NewFolderShareRepository(tx)

	created, err := repo.Create(context.Background(), folder.ID, "tok_abc123")
	require.NoError(t, err)
	assert.NotEqual(t, "", created.ID.String())

	share, err := repo.GetByFolderID(context.Background(), folder.ID)

	require.NoError(t, err)
	assert.Equal(t, folder.ID, share.FolderID)
	assert.Equal(t, "tok_abc123", share.Token)
}

func TestFolderShareRepository_GetByFolderID_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	repo := NewFolderShareRepository(tx)

	_, err := repo.GetByFolderID(context.Background(), folder.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestFolderShareRepository_Create_DuplicateFolderID(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	repo := NewFolderShareRepository(tx)

	_, err := repo.Create(context.Background(), folder.ID, "tok_first")
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), folder.ID, "tok_second")
	require.ErrorContains(t, err, "23505")
}

// --- GetByToken ---

func TestFolderShareRepository_GetByToken_PreloadsFolder(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	description := "trip board"
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	require.NoError(t, tx.Model(folder).Update("description", description).Error)
	repo := NewFolderShareRepository(tx)

	_, err := repo.Create(context.Background(), folder.ID, "tok_abc123")
	require.NoError(t, err)

	share, err := repo.GetByToken(context.Background(), "tok_abc123")

	require.NoError(t, err)
	assert.Equal(t, "travel", share.Folder.Name)
	require.NotNil(t, share.Folder.Description)
	assert.Equal(t, description, *share.Folder.Description)
	assert.Equal(t, mustMakeUserID("kp_abc123"), share.Folder.UserID)
}

func TestFolderShareRepository_GetByToken_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewFolderShareRepository(tx)

	_, err := repo.GetByToken(context.Background(), "unknown-token")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// --- DeleteByFolderID ---

func TestFolderShareRepository_DeleteByFolderID_RemovesRow(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	repo := NewFolderShareRepository(tx)

	_, err := repo.Create(context.Background(), folder.ID, "tok_abc123")
	require.NoError(t, err)

	require.NoError(t, repo.DeleteByFolderID(context.Background(), folder.ID))

	_, err = repo.GetByFolderID(context.Background(), folder.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestFolderShareRepository_DeleteByFolderID_NoRowIsNoop(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	repo := NewFolderShareRepository(tx)

	err := repo.DeleteByFolderID(context.Background(), folder.ID)

	require.NoError(t, err)
}

// --- Cascade delete ---

func TestFolderShareRepository_CascadeDelete_WhenFolderDeleted(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_abc123")
	folder := createFolder(t, tx, mustMakeUserID("kp_abc123"),"travel", nil)
	repo := NewFolderShareRepository(tx)

	_, err := repo.Create(context.Background(), folder.ID, "tok_abc123")
	require.NoError(t, err)

	require.NoError(t, NewFolderRepository(tx).DeleteWithCascade(context.Background(), folder.ID, mustMakeUserID("kp_abc123")))

	_, err = repo.GetByFolderID(context.Background(), folder.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
