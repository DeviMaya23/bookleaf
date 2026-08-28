package repository

import (
	"context"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/testutil"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- ImageRepository: ListAllByUserID / HardDeleteAllByUserID ---

func TestImageRepository_ListAllByUserID_IncludesTrashedExcludesOtherUsers(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_imgall_owner")
	createUser(t, tx, "kp_imgall_other")
	ownerID := mustMakeUserID("kp_imgall_owner")
	repo := NewImageRepository(tx)

	active, err := repo.Create(context.Background(), newTestImage(ownerID))
	require.NoError(t, err)
	trashed, err := repo.Create(context.Background(), newTestImage(ownerID))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDelete(context.Background(), trashed.ID, ownerID))
	_, err = repo.Create(context.Background(), newTestImage(mustMakeUserID("kp_imgall_other")))
	require.NoError(t, err)

	images, err := repo.ListAllByUserID(context.Background(), ownerID)

	require.NoError(t, err)
	ids := []uuid.UUID{images[0].ID, images[1].ID}
	assert.ElementsMatch(t, []uuid.UUID{active.ID, trashed.ID}, ids)
}

func TestImageRepository_HardDeleteAllByUserID_RemovesAllIncludingTrashedLeavesOtherUsers(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_imghard_owner")
	createUser(t, tx, "kp_imghard_other")
	ownerID := mustMakeUserID("kp_imghard_owner")
	otherID := mustMakeUserID("kp_imghard_other")
	repo := NewImageRepository(tx)

	_, err := repo.Create(context.Background(), newTestImage(ownerID))
	require.NoError(t, err)
	trashed, err := repo.Create(context.Background(), newTestImage(ownerID))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDelete(context.Background(), trashed.ID, ownerID))
	otherImg, err := repo.Create(context.Background(), newTestImage(otherID))
	require.NoError(t, err)

	err = repo.HardDeleteAllByUserID(context.Background(), ownerID)

	require.NoError(t, err)
	var ownerCount int64
	tx.Unscoped().Model(&domain.Image{}).Where("user_id = ?", ownerID).Count(&ownerCount)
	assert.EqualValues(t, 0, ownerCount)

	var got domain.Image
	require.NoError(t, tx.Where("id = ?", otherImg.ID).First(&got).Error)
}

// --- FolderRepository: ClearAllParents / DeleteAllByUserID ---

func TestFolderRepository_ClearAllParents_NullsOwnFoldersOnly(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_folderclear_owner")
	createUser(t, tx, "kp_folderclear_other")
	ownerID := mustMakeUserID("kp_folderclear_owner")
	otherID := mustMakeUserID("kp_folderclear_other")
	parent := createFolder(t, tx, ownerID, "parent", nil)
	child := createFolder(t, tx, ownerID, "child", &parent.ID)
	otherParent := createFolder(t, tx, otherID, "other-parent", nil)
	otherChild := createFolder(t, tx, otherID, "other-child", &otherParent.ID)
	repo := NewFolderRepository(tx)

	err := repo.ClearAllParents(context.Background(), ownerID)

	require.NoError(t, err)
	var gotParent, gotChild, gotOtherChild domain.Folder
	require.NoError(t, tx.Where("id = ?", parent.ID).First(&gotParent).Error)
	require.NoError(t, tx.Where("id = ?", child.ID).First(&gotChild).Error)
	require.NoError(t, tx.Where("id = ?", otherChild.ID).First(&gotOtherChild).Error)
	assert.Nil(t, gotChild.ParentID)
	require.NotNil(t, gotOtherChild.ParentID)
	assert.Equal(t, otherParent.ID, *gotOtherChild.ParentID)
}

func TestFolderRepository_DeleteAllByUserID_RemovesOwnFoldersLeavesOtherUsers(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_folderdel_owner")
	createUser(t, tx, "kp_folderdel_other")
	ownerID := mustMakeUserID("kp_folderdel_owner")
	otherID := mustMakeUserID("kp_folderdel_other")
	parent := createFolder(t, tx, ownerID, "parent", nil)
	createFolder(t, tx, ownerID, "child", &parent.ID)
	otherFolder := createFolder(t, tx, otherID, "folder", nil)
	repo := NewFolderRepository(tx)
	require.NoError(t, repo.ClearAllParents(context.Background(), ownerID))

	err := repo.DeleteAllByUserID(context.Background(), ownerID)

	require.NoError(t, err)
	var ownerCount int64
	tx.Model(&domain.Folder{}).Where("user_id = ?", ownerID).Count(&ownerCount)
	assert.EqualValues(t, 0, ownerCount)

	var got domain.Folder
	require.NoError(t, tx.Where("id = ?", otherFolder.ID).First(&got).Error)
}

// --- TagRepository: DeleteAllByUserID ---

func TestTagRepository_DeleteAllByUserID_RemovesOwnTagsCascadesImageTagsLeavesOtherUsers(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_tagdel_owner")
	createUser(t, tx, "kp_tagdel_other")
	ownerID := mustMakeUserID("kp_tagdel_owner")
	otherID := mustMakeUserID("kp_tagdel_other")
	imageRepo := NewImageRepository(tx)
	tagRepo := NewTagRepository(tx)

	img, err := imageRepo.Create(context.Background(), newTestImage(ownerID))
	require.NoError(t, err)
	tag, err := tagRepo.Create(context.Background(), newTestTag(ownerID, "wipetag"))
	require.NoError(t, err)
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag.ID}))
	otherTag, err := tagRepo.Create(context.Background(), newTestTag(otherID, "othertag"))
	require.NoError(t, err)

	err = tagRepo.DeleteAllByUserID(context.Background(), ownerID)

	require.NoError(t, err)
	var ownerTagCount, imageTagCount int64
	tx.Model(&domain.Tag{}).Where("user_id = ?", ownerID).Count(&ownerTagCount)
	tx.Table("image_tags").Where("image_id = ?", img.ID).Count(&imageTagCount)
	assert.EqualValues(t, 0, ownerTagCount)
	assert.EqualValues(t, 0, imageTagCount)

	var got domain.Tag
	require.NoError(t, tx.Where("id = ?", otherTag.ID).First(&got).Error)
}

// --- PendingUploadRepository: ListAllByUserID / DeleteAllByUserID ---

func TestPendingUploadRepository_ListAllAndDeleteAllByUserID_ScopedToOwner(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_pendall_owner")
	createUser(t, tx, "kp_pendall_other")
	ownerID := mustMakeUserID("kp_pendall_owner")
	otherID := mustMakeUserID("kp_pendall_other")
	repo := NewPendingUploadRepository(tx)

	owned, err := repo.Create(context.Background(), &domain.PendingUpload{
		UserID: ownerID, Title: "owned", R2Path: "users/" + ownerID.String() + "/images/1.jpg", MIMEType: "image/jpeg",
	})
	require.NoError(t, err)
	other, err := repo.Create(context.Background(), &domain.PendingUpload{
		UserID: otherID, Title: "other", R2Path: "users/" + otherID.String() + "/images/1.jpg", MIMEType: "image/jpeg",
	})
	require.NoError(t, err)

	listed, err := repo.ListAllByUserID(context.Background(), ownerID)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, owned.ID, listed[0].ID)

	err = repo.DeleteAllByUserID(context.Background(), ownerID)

	require.NoError(t, err)
	remaining, err := repo.ListAllByUserID(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Empty(t, remaining)

	gotOther, err := repo.GetByID(context.Background(), other.ID, otherID)
	require.NoError(t, err)
	assert.Equal(t, other.ID, gotOther.ID)
}

// --- UserRepository: SetAccountState / MarkPurged / ListByAccountState / ListPurgedBefore / HardDelete ---

func TestUserRepository_SetAccountState_UpdatesState(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_setstate")
	ownerID := mustMakeUserID("kp_setstate")
	repo := NewUserRepository(tx)

	err := repo.SetAccountState(context.Background(), ownerID, domain.AccountStatePendingDeletion)

	require.NoError(t, err)
	got, err := repo.GetByID(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatePendingDeletion, got.AccountState)
}

func TestUserRepository_SetAccountState_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	err := repo.SetAccountState(context.Background(), uuid.New(), domain.AccountStatePendingDeletion)

	require.ErrorIs(t, err, usecase.ErrUserNotFound)
}

func TestUserRepository_MarkPurged_SetsStateAndTimestamp(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_markpurged")
	ownerID := mustMakeUserID("kp_markpurged")
	repo := NewUserRepository(tx)
	purgedAt := time.Now().Truncate(time.Millisecond)

	err := repo.MarkPurged(context.Background(), ownerID, purgedAt)

	require.NoError(t, err)
	got, err := repo.GetByID(context.Background(), ownerID)
	require.NoError(t, err)
	assert.Equal(t, domain.AccountStatePurged, got.AccountState)
	require.NotNil(t, got.PurgedAt)
	assert.WithinDuration(t, purgedAt, *got.PurgedAt, time.Second)
}

func TestUserRepository_ListByAccountState_ReturnsMatchingUsers(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_bystate_pending")
	createUser(t, tx, "kp_bystate_active")
	pendingID := mustMakeUserID("kp_bystate_pending")
	activeID := mustMakeUserID("kp_bystate_active")
	repo := NewUserRepository(tx)
	require.NoError(t, repo.SetAccountState(context.Background(), pendingID, domain.AccountStatePendingDeletion))

	users, err := repo.ListByAccountState(context.Background(), domain.AccountStatePendingDeletion)

	require.NoError(t, err)
	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	assert.Contains(t, ids, pendingID)
	assert.NotContains(t, ids, activeID)
}

func TestUserRepository_ListByAccountState_EmptyWhenNoneMatch(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	users, err := repo.ListByAccountState(context.Background(), domain.AccountStatePurged)

	require.NoError(t, err)
	for _, u := range users {
		assert.NotEqual(t, domain.AccountStateActive, u.AccountState)
	}
}

func TestUserRepository_ListPurgedBefore_ReturnsExpiredTombstones(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_purgedbefore_old")
	createUser(t, tx, "kp_purgedbefore_recent")
	oldID := mustMakeUserID("kp_purgedbefore_old")
	recentID := mustMakeUserID("kp_purgedbefore_recent")
	repo := NewUserRepository(tx)

	longAgo := time.Now().Add(-48 * time.Hour)
	require.NoError(t, repo.MarkPurged(context.Background(), oldID, longAgo))
	justNow := time.Now()
	require.NoError(t, repo.MarkPurged(context.Background(), recentID, justNow))

	threshold := time.Now().Add(-25 * time.Hour)
	users, err := repo.ListPurgedBefore(context.Background(), threshold)

	require.NoError(t, err)
	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	assert.Contains(t, ids, oldID)
	assert.NotContains(t, ids, recentID)
}

func TestUserRepository_HardDelete_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_harddeluser")
	ownerID := mustMakeUserID("kp_harddeluser")
	repo := NewUserRepository(tx)

	err := repo.HardDelete(context.Background(), ownerID)

	require.NoError(t, err)
	var count int64
	tx.Unscoped().Model(&domain.User{}).Where("id = ?", ownerID).Count(&count)
	assert.EqualValues(t, 0, count)
}

func TestUserRepository_HardDelete_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	repo := NewUserRepository(tx)

	err := repo.HardDelete(context.Background(), uuid.New())

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// --- AccountRepository.Transaction: full account wipe ---

func TestAccountRepository_Transaction_WipesAllUserDataInFKOrder(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	createUser(t, tx, "kp_accountwipe")
	userID := mustMakeUserID("kp_accountwipe")

	parent := createFolder(t, tx, userID, "parent", nil)
	child := createFolder(t, tx, userID, "child", &parent.ID)

	imageRepo := NewImageRepository(tx)
	activeImg, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, imageRepo.SetImageFolder(context.Background(), activeImg.ID, &child.ID))
	trashedImg, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, imageRepo.SoftDelete(context.Background(), trashedImg.ID, userID))

	tagRepo := NewTagRepository(tx)
	tag, err := tagRepo.Create(context.Background(), newTestTag(userID, "wipetag"))
	require.NoError(t, err)
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), activeImg.ID, []uuid.UUID{tag.ID}))

	pendingRepo := NewPendingUploadRepository(tx)
	_, err = pendingRepo.Create(context.Background(), &domain.PendingUpload{
		UserID: userID, Title: "pending", R2Path: "users/" + userID.String() + "/images/pending.jpg", MIMEType: "image/jpeg",
	})
	require.NoError(t, err)

	accountRepo := NewAccountRepository(tx)

	var collectedImages []*domain.Image
	var collectedPending []*domain.PendingUpload
	err = accountRepo.Transaction(context.Background(), func(repos usecase.AccountRepos) error {
		if err := repos.Folders.ClearAllParents(context.Background(), userID); err != nil {
			return err
		}

		images, err := repos.Images.ListAllByUserID(context.Background(), userID)
		if err != nil {
			return err
		}
		collectedImages = images
		if err := repos.Images.HardDeleteAllByUserID(context.Background(), userID); err != nil {
			return err
		}

		if err := repos.Folders.DeleteAllByUserID(context.Background(), userID); err != nil {
			return err
		}

		if err := repos.Tags.DeleteAllByUserID(context.Background(), userID); err != nil {
			return err
		}

		pendingUploads, err := repos.PendingUploads.ListAllByUserID(context.Background(), userID)
		if err != nil {
			return err
		}
		collectedPending = pendingUploads
		return repos.PendingUploads.DeleteAllByUserID(context.Background(), userID)
	})

	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{activeImg.ID, trashedImg.ID}, []uuid.UUID{collectedImages[0].ID, collectedImages[1].ID})
	require.Len(t, collectedPending, 1)

	var folderCount, imageCount, tagCount, pendingCount, imageFolderCount, imageTagCount int64
	tx.Model(&domain.Folder{}).Where("user_id = ?", userID).Count(&folderCount)
	tx.Unscoped().Model(&domain.Image{}).Where("user_id = ?", userID).Count(&imageCount)
	tx.Model(&domain.Tag{}).Where("user_id = ?", userID).Count(&tagCount)
	tx.Model(&domain.PendingUpload{}).Where("user_id = ?", userID).Count(&pendingCount)
	tx.Table("image_folders").Where("image_id = ?", activeImg.ID).Count(&imageFolderCount)
	tx.Table("image_tags").Where("image_id = ?", activeImg.ID).Count(&imageTagCount)

	assert.EqualValues(t, 0, folderCount)
	assert.EqualValues(t, 0, imageCount)
	assert.EqualValues(t, 0, tagCount)
	assert.EqualValues(t, 0, pendingCount)
	assert.EqualValues(t, 0, imageFolderCount)
	assert.EqualValues(t, 0, imageTagCount)

	// The users row is NOT touched by the wipe transaction
	var gotUser domain.User
	require.NoError(t, tx.Where("id = ?", userID).First(&gotUser).Error)
	assert.Equal(t, domain.AccountStateActive, gotUser.AccountState)
}
