package repository

import (
	"context"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPendingUploadRepository_Create_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_pending_create")
	repo := NewPendingUploadRepository(tx)

	description := "cover image"
	sourceURL := "https://example.com"
	pending, err := repo.Create(context.Background(), &domain.PendingUpload{
		UserID:      "kp_pending_create",
		Title:       "sunset",
		Description: &description,
		SourceURL:   &sourceURL,
		R2Path:      "users/kp_pending_create/images/test.jpg",
		MIMEType:    "image/jpeg",
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, pending.ID)

	var found domain.PendingUpload
	require.NoError(t, tx.Where("id = ?", pending.ID).First(&found).Error)
	assert.Equal(t, "kp_pending_create", found.UserID)
	assert.Equal(t, "sunset", found.Title)
	require.NotNil(t, found.Description)
	assert.Equal(t, description, *found.Description)
	require.NotNil(t, found.SourceURL)
	assert.Equal(t, sourceURL, *found.SourceURL)
	assert.Equal(t, "users/kp_pending_create/images/test.jpg", found.R2Path)
	assert.Equal(t, "image/jpeg", found.MIMEType)
}

func TestPendingUploadRepository_GetByID_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_pending_get")
	repo := NewPendingUploadRepository(tx)

	pending, err := repo.Create(context.Background(), &domain.PendingUpload{
		UserID:   "kp_pending_get",
		Title:    "beach",
		R2Path:   "users/kp_pending_get/images/test.jpg",
		MIMEType: "image/jpeg",
	})
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), pending.ID, "kp_pending_get")

	require.NoError(t, err)
	assert.Equal(t, pending.ID, found.ID)
	assert.Equal(t, pending.UserID, found.UserID)
	assert.Equal(t, pending.Title, found.Title)
}

func TestPendingUploadRepository_GetByID_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewPendingUploadRepository(tx)

	_, err := repo.GetByID(context.Background(), uuid.New(), "kp_missing")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestPendingUploadRepository_GetByID_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_pend_owner")
	createUser(t, tx, "kp_pend_other")
	repo := NewPendingUploadRepository(tx)

	pending, err := repo.Create(context.Background(), &domain.PendingUpload{
		UserID:   "kp_pend_owner",
		Title:    "private",
		R2Path:   "users/kp_pend_owner/images/test.jpg",
		MIMEType: "image/jpeg",
	})
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), pending.ID, "kp_pend_other")

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestPendingUploadRepository_Delete_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_pending_delete")
	repo := NewPendingUploadRepository(tx)

	pending, err := repo.Create(context.Background(), &domain.PendingUpload{
		UserID:   "kp_pending_delete",
		Title:    "mountain",
		R2Path:   "users/kp_pending_delete/images/test.jpg",
		MIMEType: "image/jpeg",
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(context.Background(), pending.ID))

	var found domain.PendingUpload
	err = tx.Where("id = ?", pending.ID).First(&found).Error
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestPendingUploadRepository_ListStale(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_pending_stale")
	repo := NewPendingUploadRepository(tx)

	oldCreatedAt := time.Now().Add(-2 * time.Hour)
	newCreatedAt := time.Now().Add(-5 * time.Minute)

	old := &domain.PendingUpload{
		ID:        uuid.New(),
		UserID:    "kp_pending_stale",
		Title:     "old",
		R2Path:    "users/kp_pending_stale/images/old.jpg",
		MIMEType:  "image/jpeg",
		CreatedAt: oldCreatedAt,
	}
	newer := &domain.PendingUpload{
		ID:        uuid.New(),
		UserID:    "kp_pending_stale",
		Title:     "new",
		R2Path:    "users/kp_pending_stale/images/new.jpg",
		MIMEType:  "image/jpeg",
		CreatedAt: newCreatedAt,
	}
	require.NoError(t, tx.Create(old).Error)
	require.NoError(t, tx.Create(newer).Error)

	stale, err := repo.ListStale(context.Background(), time.Now().Add(-30*time.Minute))

	require.NoError(t, err)
	require.Len(t, stale, 1)
	assert.Equal(t, old.ID, stale[0].ID)
}
