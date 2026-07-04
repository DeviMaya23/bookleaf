package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

func setupImageTest(t *testing.T) (*imageRepository, string) {
	t.Helper()
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_imgtest")
	require.NoError(t, err)

	return NewImageRepository(tx), user.ID
}

func newTestImage(userID string) *domain.Image {
	id := uuid.New()
	return &domain.Image{
		ID:         id,
		UserID:     userID,
		Title:      "test image",
		MIMEType:   "image/jpeg",
		R2Path:     "users/" + userID + "/images/" + id.String() + ".jpg",
	}
}

func TestImageRepository_Create_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	img, err := repo.Create(context.Background(), newTestImage(userID))

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, img.ID)
	assert.Equal(t, userID, img.UserID)
	assert.Equal(t, "test image", img.Title)
}


func TestImageRepository_List_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), userID, false, nil, nil, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Len(t, images, 2)
}

func TestImageRepository_List_ExcludesOtherUserImages(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown7")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth7")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	_, err = repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), other.ID, false, nil, nil, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestImageRepository_List_FilterByFolderIDs_SingleValue(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_imgtest")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "moodboard"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	img1, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img1.ID, &folder.ID))

	_, err = repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), user.ID, false, []uuid.UUID{folder.ID}, nil, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Len(t, images, 1)
	require.Len(t, images[0].ImageFolders, 1)
	assert.Equal(t, folder.ID, images[0].ImageFolders[0].FolderID)
}

func TestImageRepository_List_FilterByFolderIDs_MatchAny(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_folderany")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folderA, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "folder-a"})
	require.NoError(t, err)
	folderB, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "folder-b"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	inA, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), inA.ID, &folderA.ID))

	inBoth, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SyncImageFolders(context.Background(), inBoth.ID, []uuid.UUID{folderA.ID, folderB.ID}))

	_, err = repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), user.ID, false, []uuid.UUID{folderA.ID, folderB.ID}, nil, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	require.Len(t, images, 2, "inBoth must appear exactly once despite matching both supplied folder IDs")
	assert.ElementsMatch(t, []uuid.UUID{inA.ID, inBoth.ID}, []uuid.UUID{images[0].ID, images[1].ID})
}

func TestImageRepository_GetByID_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestImageRepository_GetByID_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.GetByID(context.Background(), uuid.New(), userID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_GetByID_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown1")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth1")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	created, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, other.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_UpdateThumbnailPath_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	thumbPath := "users/" + userID + "/thumbnails/" + created.ID.String() + ".jpg"
	err = repo.UpdateThumbnailPath(context.Background(), created.ID, thumbPath)

	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, found.ThumbnailPath)
	assert.Equal(t, thumbPath, *found.ThumbnailPath)
}

func TestImageRepository_UpdateThumbnailPath_NotFound(t *testing.T) {
	repo, _ := setupImageTest(t)

	err := repo.UpdateThumbnailPath(context.Background(), uuid.New(), "some/path.jpg")

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}


func TestImageRepository_ListUnlabelled_ReturnsOnlyUnlabelledNonDeleted(t *testing.T) {
	repo, userID := setupImageTest(t)

	unlabelled, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	labelled, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, repo.UpdateLabels(context.Background(), labelled.ID, json.RawMessage(`[]`), nil))

	deleted, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDelete(context.Background(), deleted.ID, userID))

	images, err := repo.ListUnlabelled(context.Background(), userID)

	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, unlabelled.ID, images[0].ID)
}

func TestImageRepository_ListUnlabelled_ExcludesOtherUserImages(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown8")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth8")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	_, err = repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)

	images, err := repo.ListUnlabelled(context.Background(), other.ID)

	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestImageRepository_ListUnlabelled_Empty(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, repo.UpdateLabels(context.Background(), created.ID, json.RawMessage(`[]`), nil))

	images, err := repo.ListUnlabelled(context.Background(), userID)

	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestImageRepository_SoftDelete_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), created.ID, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_SoftDelete_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	err := repo.SoftDelete(context.Background(), uuid.New(), userID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_SoftDelete_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown2")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth2")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	created, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, other.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_GetDeletedByID_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	deleted, err := repo.GetDeletedByID(context.Background(), created.ID, userID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, deleted.ID)
	assert.True(t, deleted.DeletedAt.Valid)
}

func TestImageRepository_GetDeletedByID_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	// non-deleted image should not be returned by GetDeletedByID
	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	_, err = repo.GetDeletedByID(context.Background(), created.ID, userID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_GetDeletedByID_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown3")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth3")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	created, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDelete(context.Background(), created.ID, owner.ID))

	_, err = repo.GetDeletedByID(context.Background(), created.ID, other.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_Restore_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	created, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	err = repo.Restore(context.Background(), created.ID, userID)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), created.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.False(t, found.DeletedAt.Valid)
}

func TestImageRepository_Restore_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	err := repo.Restore(context.Background(), uuid.New(), userID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_Restore_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown4")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth4")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	created, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDelete(context.Background(), created.ID, owner.ID))

	err = repo.Restore(context.Background(), created.ID, other.ID)

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_ListTrashed_Success(t *testing.T) {
	repo, userID := setupImageTest(t)

	img1, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	img2, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	err = repo.SoftDelete(context.Background(), img1.ID, userID)
	require.NoError(t, err)
	err = repo.SoftDelete(context.Background(), img2.ID, userID)
	require.NoError(t, err)

	// create one non-deleted image to confirm it's excluded
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	trashed, err := repo.ListTrashed(context.Background(), userID, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Len(t, trashed, 2)
}

func TestImageRepository_ListTrashed_Empty(t *testing.T) {
	repo, userID := setupImageTest(t)

	trashed, err := repo.ListTrashed(context.Background(), userID, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Empty(t, trashed)
}

func TestImageRepository_ListTrashed_ExcludesOtherUserImages(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown5")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth5")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDelete(context.Background(), img.ID, owner.ID))

	trashed, err := repo.ListTrashed(context.Background(), other.ID, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Empty(t, trashed)
}

func TestImageRepository_List_Pagination_FirstPage(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	// limit=2 → repo should return limit+1=3 rows, signalling more data exists
	images, err := repo.List(context.Background(), userID, false, nil, nil, nil, nil, nil, nil, nil, 2)

	require.NoError(t, err)
	assert.Len(t, images, 3)
}

func TestImageRepository_List_Pagination_WithCursor(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	second, err := repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	// first page: returns 3 rows (limit+1), ordered created_at DESC, id DESC
	firstPage, err := repo.List(context.Background(), userID, false, nil, nil, nil, nil, nil, nil, nil, 2)
	require.NoError(t, err)
	require.Len(t, firstPage, 3)

	// cursor from the 2nd item (last of the visible page)
	cursorItem := firstPage[1]
	cursor := &usecase.ImageCursor{CreatedAt: cursorItem.CreatedAt, ID: cursorItem.ID}

	// second page: should return only items after the cursor
	secondPage, err := repo.List(context.Background(), userID, false, nil, nil, nil, nil, nil, nil, cursor, 2)

	require.NoError(t, err)
	assert.Len(t, secondPage, 1)

	// IDs on page 2 must not overlap with page 1
	page1IDs := map[uuid.UUID]bool{firstPage[0].ID: true, firstPage[1].ID: true}
	for _, img := range secondPage {
		assert.False(t, page1IDs[img.ID], "page 2 item %s appeared on page 1", img.ID)
	}
	_ = second
}

func TestImageRepository_ListTrashed_Pagination_FirstPage(t *testing.T) {
	repo, userID := setupImageTest(t)

	for range 3 {
		img, err := repo.Create(context.Background(), newTestImage(userID))
		require.NoError(t, err)
		require.NoError(t, repo.SoftDelete(context.Background(), img.ID, userID))
	}

	// limit=2 → repo should return limit+1=3 rows
	trashed, err := repo.ListTrashed(context.Background(), userID, nil, nil, nil, nil, 2)

	require.NoError(t, err)
	assert.Len(t, trashed, 3)
}

func TestImageRepository_ListTrashed_Pagination_WithCursor(t *testing.T) {
	repo, userID := setupImageTest(t)

	for range 3 {
		img, err := repo.Create(context.Background(), newTestImage(userID))
		require.NoError(t, err)
		require.NoError(t, repo.SoftDelete(context.Background(), img.ID, userID))
	}

	// first page (limit=2, returns 3 since repo returns limit+1 to signal more data)
	firstPage, err := repo.ListTrashed(context.Background(), userID, nil, nil, nil, nil, 2)
	require.NoError(t, err)
	require.Len(t, firstPage, 3)

	cursorItem := firstPage[1]
	cursor := &usecase.ImageCursor{CreatedAt: cursorItem.CreatedAt, ID: cursorItem.ID}

	// second page
	secondPage, err := repo.ListTrashed(context.Background(), userID, nil, nil, nil, cursor, 2)

	require.NoError(t, err)
	assert.Len(t, secondPage, 1)

	page1IDs := map[uuid.UUID]bool{firstPage[0].ID: true, firstPage[1].ID: true}
	for _, img := range secondPage {
		assert.False(t, page1IDs[img.ID], "page 2 item %s appeared on page 1", img.ID)
	}
}

func TestImageRepository_ListTrashed_DefaultsToCreatedAtDesc(t *testing.T) {
	repo, userID := setupImageTest(t)

	for range 3 {
		img, err := repo.Create(context.Background(), newTestImage(userID))
		require.NoError(t, err)
		require.NoError(t, repo.SoftDelete(context.Background(), img.ID, userID))
	}

	trashed, err := repo.ListTrashed(context.Background(), userID, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	require.Len(t, trashed, 3)
	for i := 1; i < len(trashed); i++ {
		prev := trashed[i-1].CreatedAt
		curr := trashed[i].CreatedAt
		assert.True(t, !curr.After(prev), "expected created_at DESC: item %d (%s) is after item %d (%s)", i, curr, i-1, prev)
	}
}

func TestImageRepository_ListTrashed_Pagination_SortAware(t *testing.T) {
	tests := []struct {
		name      string
		sortField string
		direction string
		ascending bool // true if creation order (index 0..n-1) matches the expected page order
	}{
		{name: "title ascending", sortField: "title", direction: "asc", ascending: true},
		{name: "title descending", sortField: "title", direction: "desc", ascending: false},
		{name: "created_at ascending", sortField: "created_at", direction: "asc", ascending: true},
		{name: "created_at descending", sortField: "created_at", direction: "desc", ascending: false},
		{name: "deleted_at ascending", sortField: "deleted_at", direction: "asc", ascending: true},
		{name: "deleted_at descending", sortField: "deleted_at", direction: "desc", ascending: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.NewTestTx(t, testDB)
			userRepo := NewUserRepository(tx)
			user, err := userRepo.GetOrCreate(context.Background(), "kp_trashsortpage")
			require.NoError(t, err)
			repo := NewImageRepository(tx)

			base := time.Now().Add(-time.Hour)
			created := make([]*domain.Image, 0, 5)
			for i := range 5 {
				img := newTestImage(user.ID)
				img.Title = fmt.Sprintf("image-%d", i)
				img.CreatedAt = base.Add(time.Duration(i) * time.Minute)
				saved, err := repo.Create(context.Background(), img)
				require.NoError(t, err)
				require.NoError(t, repo.SoftDelete(context.Background(), saved.ID, user.ID))
				created = append(created, saved)
			}

			expected := make([]uuid.UUID, len(created))
			for i, img := range created {
				if tt.ascending {
					expected[i] = img.ID
				} else {
					expected[len(created)-1-i] = img.ID
				}
			}

			collected := listAllTrashedPages(t, repo, user.ID, tt.sortField, tt.direction, 2)

			assert.Equal(t, expected, collected)
		})
	}
}

func TestImageRepository_Update_SelectiveFieldUpdate(t *testing.T) {
	repo, userID := setupImageTest(t)

	img := newTestImage(userID)
	thumbPath := "users/" + userID + "/thumbnails/test.jpg"
	img.ThumbnailPath = &thumbPath
	created, err := repo.Create(context.Background(), img)
	require.NoError(t, err)

	updated, err := repo.Update(context.Background(), created.ID, userID, map[string]any{
		"title": "new title",
	})

	require.NoError(t, err)
	assert.Equal(t, "new title", updated.Title)
	require.NotNil(t, updated.ThumbnailPath)
	assert.Equal(t, thumbPath, *updated.ThumbnailPath)
}

func TestImageRepository_Update_NotFound(t *testing.T) {
	repo, userID := setupImageTest(t)

	_, err := repo.Update(context.Background(), uuid.New(), userID, map[string]any{
		"title": "ghost",
	})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_Update_WrongUser(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_imgown6")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_imgoth6")
	require.NoError(t, err)
	repo := NewImageRepository(tx)
	created, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)

	_, err = repo.Update(context.Background(), created.ID, other.ID, map[string]any{"title": "stolen"})

	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_CountByFolderID(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_imgcount")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "count-target"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	img1, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img1.ID, &folder.ID))

	img2, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img2.ID, &folder.ID))

	otherFolder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "other"})
	require.NoError(t, err)
	img3, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img3.ID, &otherFolder.ID))

	// soft-delete img2 — should be excluded from count
	require.NoError(t, repo.SoftDelete(context.Background(), img2.ID, user.ID))

	count, err := repo.CountByFolderID(context.Background(), folder.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestImageRepository_SetImageFolder_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_setfolder")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "assigned"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	err = repo.SetImageFolder(context.Background(), img.ID, &folder.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), img.ID, user.ID)
	require.NoError(t, err)
	require.Len(t, found.ImageFolders, 1)
	assert.Equal(t, folder.ID, found.ImageFolders[0].FolderID)
	assert.Equal(t, "a0", found.ImageFolders[0].Position)
}

func TestImageRepository_SetImageFolder_Unfile(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_unfile")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "toremove"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	require.NoError(t, repo.SetImageFolder(context.Background(), img.ID, &folder.ID))
	require.NoError(t, repo.SetImageFolder(context.Background(), img.ID, nil))

	found, err := repo.GetByID(context.Background(), img.ID, user.ID)
	require.NoError(t, err)
	assert.Empty(t, found.ImageFolders)
}

func TestImageRepository_SetImageFolder_NoopWhenUnfilingImageWithNoFolder(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_noop")
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	err = repo.SetImageFolder(context.Background(), img.ID, nil)
	require.NoError(t, err)
}

func TestImageRepository_List_Unfiled(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_unfiled")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "assigned"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	filed, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), filed.ID, &folder.ID))

	_, err = repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), user.ID, true, nil, nil, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Len(t, images, 1)
	for _, img := range images {
		assert.Empty(t, img.ImageFolders)
	}
}

func TestImageRepository_List_PreloadsTags(t *testing.T) {
	tagRepo, imageRepo, userID := setupTagTest(t)

	img, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	tag, err := tagRepo.Create(context.Background(), newTestTag(userID, "preloadcheck"))
	require.NoError(t, err)

	err = tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag.ID})
	require.NoError(t, err)

	images, err := imageRepo.List(context.Background(), userID, false, nil, nil, nil, nil, nil, nil, nil, 200)
	require.NoError(t, err)

	var found *domain.Image
	for _, i := range images {
		if i.ID == img.ID {
			found = i
			break
		}
	}
	require.NotNil(t, found)
	require.Len(t, found.Tags, 1)
	assert.Equal(t, tag.ID, found.Tags[0].ID)
}

func TestImageRepository_List_FilterByTagIDs_SingleValue(t *testing.T) {
	tagRepo, imageRepo, userID := setupTagTest(t)

	tagged, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	_, err = imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	tag, err := tagRepo.Create(context.Background(), newTestTag(userID, "filterme"))
	require.NoError(t, err)

	err = tagRepo.ReplaceImageTags(context.Background(), tagged.ID, []uuid.UUID{tag.ID})
	require.NoError(t, err)

	images, err := imageRepo.List(context.Background(), userID, false, nil, []uuid.UUID{tag.ID}, nil, nil, nil, nil, nil, 200)
	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, tagged.ID, images[0].ID)
}

func TestImageRepository_List_FilterByTagIDs_MatchAny(t *testing.T) {
	tagRepo, imageRepo, userID := setupTagTest(t)

	tagX, err := tagRepo.Create(context.Background(), newTestTag(userID, "tag-x"))
	require.NoError(t, err)
	tagY, err := tagRepo.Create(context.Background(), newTestTag(userID, "tag-y"))
	require.NoError(t, err)

	withX, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), withX.ID, []uuid.UUID{tagX.ID}))

	withBoth, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), withBoth.ID, []uuid.UUID{tagX.ID, tagY.ID}))

	_, err = imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	images, err := imageRepo.List(context.Background(), userID, false, nil, []uuid.UUID{tagX.ID, tagY.ID}, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	require.Len(t, images, 2, "withBoth must appear exactly once despite matching both supplied tag IDs")
	assert.ElementsMatch(t, []uuid.UUID{withX.ID, withBoth.ID}, []uuid.UUID{images[0].ID, images[1].ID})
}

func TestImageRepository_List_FilterByMIMETypes_MatchAny(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_mimefilter")
	require.NoError(t, err)
	repo := NewImageRepository(tx)

	jpeg := newTestImage(user.ID)
	jpeg.MIMEType = "image/jpeg"
	jpeg, err = repo.Create(context.Background(), jpeg)
	require.NoError(t, err)

	png := newTestImage(user.ID)
	png.MIMEType = "image/png"
	png, err = repo.Create(context.Background(), png)
	require.NoError(t, err)

	gif := newTestImage(user.ID)
	gif.MIMEType = "image/gif"
	_, err = repo.Create(context.Background(), gif)
	require.NoError(t, err)

	images, err := repo.List(context.Background(), user.ID, false, nil, nil, []string{"image/jpeg", "image/png"}, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	require.Len(t, images, 2)
	assert.ElementsMatch(t, []uuid.UUID{jpeg.ID, png.ID}, []uuid.UUID{images[0].ID, images[1].ID})
}

func TestImageRepository_List_ComposesFolderAndTagFiltersWithAND(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_andcompose")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "compose"})
	require.NoError(t, err)

	tagRepo := NewTagRepository(tx)
	tag, err := tagRepo.Create(context.Background(), newTestTag(user.ID, "compose-tag"))
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	// Matches both filters.
	matchesBoth, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), matchesBoth.ID, &folder.ID))
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), matchesBoth.ID, []uuid.UUID{tag.ID}))

	// In the folder but missing the tag.
	folderOnly, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), folderOnly.ID, &folder.ID))

	// Has the tag but not in the folder.
	tagOnly, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), tagOnly.ID, []uuid.UUID{tag.ID}))

	images, err := repo.List(context.Background(), user.ID, false, []uuid.UUID{folder.ID}, []uuid.UUID{tag.ID}, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, matchesBoth.ID, images[0].ID)
}

func TestImageRepository_List_UnfiledWithFolderIDs_YieldsEmptyResult(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_contradiction")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "contradiction"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	filed, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), filed.ID, &folder.ID))

	_, err = repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	images, err := repo.List(context.Background(), user.ID, true, []uuid.UUID{folder.ID}, nil, nil, nil, nil, nil, nil, 200)

	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestImageRepository_GetByID_PreloadsTags(t *testing.T) {
	tagRepo, imageRepo, userID := setupTagTest(t)

	img, err := imageRepo.Create(context.Background(), newTestImage(userID))
	require.NoError(t, err)

	tag, err := tagRepo.Create(context.Background(), newTestTag(userID, "getbyidtag"))
	require.NoError(t, err)

	err = tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag.ID})
	require.NoError(t, err)

	found, err := imageRepo.GetByID(context.Background(), img.ID, userID)
	require.NoError(t, err)
	require.Len(t, found.Tags, 1)
	assert.Equal(t, tag.ID, found.Tags[0].ID)
}

func TestImageRepository_UpdateImageFolderPosition_Success(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_updpos")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "reorder"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img.ID, &folder.ID))

	err = repo.UpdateImageFolderPosition(context.Background(), img.ID, folder.ID, "Zz")
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), img.ID, user.ID)
	require.NoError(t, err)
	require.Len(t, found.ImageFolders, 1)
	assert.Equal(t, "Zz", found.ImageFolders[0].Position)
}

func TestImageRepository_UpdateImageFolderPosition_NotFound(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_updpos_nf")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "empty"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	err = repo.UpdateImageFolderPosition(context.Background(), img.ID, folder.ID, "a0")
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestImageRepository_ListByFolder_OrderedByPosition(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_listpos")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "ordered"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	img1, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img1.ID, &folder.ID))

	img2, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img2.ID, &folder.ID))

	img3, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img3.ID, &folder.ID))

	// Manually reorder: put img3 before img1
	require.NoError(t, repo.UpdateImageFolderPosition(context.Background(), img3.ID, folder.ID, "Zz"))

	images, err := repo.ListByFolder(context.Background(), user.ID, folder.ID, nil, nil)

	require.NoError(t, err)
	require.Len(t, images, 3)
	assert.Equal(t, img3.ID, images[0].ID)
	assert.Equal(t, img1.ID, images[1].ID)
	assert.Equal(t, img2.ID, images[2].ID)
}

func TestImageRepository_ListByFolder_ScopedToFolder(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_listscoped")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folderA, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "a"})
	require.NoError(t, err)
	folderB, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "b"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	inA, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), inA.ID, &folderA.ID))

	inB, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), inB.ID, &folderB.ID))

	images, err := repo.ListByFolder(context.Background(), user.ID, folderA.ID, nil, nil)

	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, inA.ID, images[0].ID)
}

// listAllPages walks every page of the cursor-paginated (non-folder) branch for the
// given sort/direction, asserting no duplicates appear across page boundaries, and
// returns the IDs in the order returned by the repository.
func listAllPages(t *testing.T, repo *imageRepository, userID string, sortField, direction string, pageLimit int) []uuid.UUID {
	t.Helper()

	var cursor *usecase.ImageCursor
	var collected []uuid.UUID
	seen := map[uuid.UUID]bool{}

	for {
		page, err := repo.List(context.Background(), userID, false, nil, nil, nil, nil, &sortField, &direction, cursor, pageLimit)
		require.NoError(t, err)

		take := page
		hasMore := len(page) > pageLimit
		if hasMore {
			take = page[:pageLimit]
		}

		for _, img := range take {
			assert.False(t, seen[img.ID], "image %s appeared on more than one page", img.ID)
			seen[img.ID] = true
			collected = append(collected, img.ID)
		}

		if !hasMore {
			break
		}

		last := take[len(take)-1]
		switch sortField {
		case "title":
			cursor = &usecase.ImageCursor{Title: &last.Title, ID: last.ID}
		default:
			cursor = &usecase.ImageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
		}
	}

	return collected
}

func listAllTrashedPages(t *testing.T, repo *imageRepository, userID string, sortField, direction string, pageLimit int) []uuid.UUID {
	t.Helper()

	var cursor *usecase.ImageCursor
	var collected []uuid.UUID
	seen := map[uuid.UUID]bool{}

	for {
		page, err := repo.ListTrashed(context.Background(), userID, nil, &sortField, &direction, cursor, pageLimit)
		require.NoError(t, err)

		take := page
		hasMore := len(page) > pageLimit
		if hasMore {
			take = page[:pageLimit]
		}

		for _, img := range take {
			assert.False(t, seen[img.ID], "image %s appeared on more than one page", img.ID)
			seen[img.ID] = true
			collected = append(collected, img.ID)
		}

		if !hasMore {
			break
		}

		last := take[len(take)-1]
		switch sortField {
		case "title":
			cursor = &usecase.ImageCursor{Title: &last.Title, ID: last.ID}
		case "deleted_at":
			deletedAt := last.DeletedAt.Time
			cursor = &usecase.ImageCursor{DeletedAt: &deletedAt, ID: last.ID}
		default:
			cursor = &usecase.ImageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
		}
	}

	return collected
}

func TestImageRepository_List_Pagination_SortAware(t *testing.T) {
	tests := []struct {
		name      string
		sortField string
		direction string
		ascending bool // true if creation order (index 0..n-1) matches the expected page order
	}{
		{name: "title ascending", sortField: "title", direction: "asc", ascending: true},
		{name: "title descending", sortField: "title", direction: "desc", ascending: false},
		{name: "created_at ascending", sortField: "created_at", direction: "asc", ascending: true},
		{name: "created_at descending", sortField: "created_at", direction: "desc", ascending: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := testutil.NewTestTx(t, testDB)
			userRepo := NewUserRepository(tx)
			user, err := userRepo.GetOrCreate(context.Background(), "kp_sortpage")
			require.NoError(t, err)
			repo := NewImageRepository(tx)

			base := time.Now().Add(-time.Hour)
			created := make([]*domain.Image, 0, 5)
			for i := range 5 {
				img := newTestImage(user.ID)
				img.Title = fmt.Sprintf("image-%d", i)
				img.CreatedAt = base.Add(time.Duration(i) * time.Minute)
				saved, err := repo.Create(context.Background(), img)
				require.NoError(t, err)
				created = append(created, saved)
			}

			expected := make([]uuid.UUID, len(created))
			for i, img := range created {
				if tt.ascending {
					expected[i] = img.ID
				} else {
					expected[len(created)-1-i] = img.ID
				}
			}

			collected := listAllPages(t, repo, user.ID, tt.sortField, tt.direction, 2)

			assert.Equal(t, expected, collected)
		})
	}
}

func TestImageRepository_List_Pagination_TitleTiebreaksByID(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)
	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_sorttie")
	require.NoError(t, err)
	repo := NewImageRepository(tx)

	created := make([]*domain.Image, 0, 5)
	for range 5 {
		img := newTestImage(user.ID)
		img.Title = "duplicate"
		saved, err := repo.Create(context.Background(), img)
		require.NoError(t, err)
		created = append(created, saved)
	}

	// All titles are identical, so the id tiebreak (in the same direction as the
	// primary column, i.e. ascending here) determines the order.
	expected := append([]*domain.Image{}, created...)
	sort.Slice(expected, func(i, j int) bool { return expected[i].ID.String() < expected[j].ID.String() })
	expectedIDs := make([]uuid.UUID, len(expected))
	for i, img := range expected {
		expectedIDs[i] = img.ID
	}

	collected := listAllPages(t, repo, user.ID, "title", "asc", 2)

	assert.Equal(t, expectedIDs, collected)
}

func TestImageRepository_ListByFolder_ExplicitSortOverridesPosition(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_foldersort")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "sortable"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	// Added in B, A order, so position-based ordering would list B before A.
	imgB := newTestImage(user.ID)
	imgB.Title = "B"
	imgB, err = repo.Create(context.Background(), imgB)
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), imgB.ID, &folder.ID))

	imgA := newTestImage(user.ID)
	imgA.Title = "A"
	imgA, err = repo.Create(context.Background(), imgA)
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), imgA.ID, &folder.ID))

	sortField := "title"
	direction := "asc"
	images, err := repo.ListByFolder(context.Background(), user.ID, folder.ID, &sortField, &direction)

	require.NoError(t, err)
	require.Len(t, images, 2)
	assert.Equal(t, imgA.ID, images[0].ID)
	assert.Equal(t, imgB.ID, images[1].ID)
}

func TestImageRepository_ListByFolder_PreloadsTagsAndImageFolders(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_folderpreload")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "preload-folder"})
	require.NoError(t, err)

	tagRepo := NewTagRepository(tx)
	repo := NewImageRepository(tx)

	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.SetImageFolder(context.Background(), img.ID, &folder.ID))

	tag, err := tagRepo.Create(context.Background(), newTestTag(user.ID, "folderpreloadtag"))
	require.NoError(t, err)
	require.NoError(t, tagRepo.ReplaceImageTags(context.Background(), img.ID, []uuid.UUID{tag.ID}))

	images, err := repo.ListByFolder(context.Background(), user.ID, folder.ID, nil, nil)

	require.NoError(t, err)
	require.Len(t, images, 1)
	require.Len(t, images[0].Tags, 1)
	assert.Equal(t, tag.ID, images[0].Tags[0].ID)
	require.Len(t, images[0].ImageFolders, 1)
	assert.Equal(t, folder.ID, images[0].ImageFolders[0].FolderID)
}

func TestImageRepository_AddImageToFolder_InsertsWhenAbsent(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_addtofolder_insert")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "addtofolder"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	err = repo.AddImageToFolder(context.Background(), img.ID, folder.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), img.ID, user.ID)
	require.NoError(t, err)
	require.Len(t, found.ImageFolders, 1)
	assert.Equal(t, folder.ID, found.ImageFolders[0].FolderID)
}

func TestImageRepository_AddImageToFolder_NoopWhenAlreadyPresent(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_addtofolder_noop")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "addtofolder-noop"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	img, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)

	require.NoError(t, repo.AddImageToFolder(context.Background(), img.ID, folder.ID))
	originalPosition := mustGetImageFolderPosition(t, repo, img.ID, user.ID, folder.ID)

	err = repo.AddImageToFolder(context.Background(), img.ID, folder.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), img.ID, user.ID)
	require.NoError(t, err)
	require.Len(t, found.ImageFolders, 1)
	assert.Equal(t, originalPosition, found.ImageFolders[0].Position)
}

func TestImageRepository_AddImageToFolder_AppendsAfterExistingMax(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_addtofolder_append")
	require.NoError(t, err)

	folderRepo := NewFolderRepository(tx)
	folder, err := folderRepo.Create(context.Background(), &domain.Folder{UserID: user.ID, Name: "addtofolder-append"})
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	first, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.AddImageToFolder(context.Background(), first.ID, folder.ID))
	firstPosition := mustGetImageFolderPosition(t, repo, first.ID, user.ID, folder.ID)

	second, err := repo.Create(context.Background(), newTestImage(user.ID))
	require.NoError(t, err)
	require.NoError(t, repo.AddImageToFolder(context.Background(), second.ID, folder.ID))
	secondPosition := mustGetImageFolderPosition(t, repo, second.ID, user.ID, folder.ID)

	assert.Greater(t, secondPosition, firstPosition)
}

func mustGetImageFolderPosition(t *testing.T, repo *imageRepository, imageID uuid.UUID, userID string, folderID uuid.UUID) string {
	t.Helper()
	found, err := repo.GetByID(context.Background(), imageID, userID)
	require.NoError(t, err)
	for _, f := range found.ImageFolders {
		if f.FolderID == folderID {
			return f.Position
		}
	}
	t.Fatalf("image %s not found in folder %s", imageID, folderID)
	return ""
}

func TestImageRepository_FilterOwnedImageIDs_ReturnsOnlyOwnedAndExisting(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	owner, err := userRepo.GetOrCreate(context.Background(), "kp_filterowned_owner")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_filterowned_other")
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	owned, err := repo.Create(context.Background(), newTestImage(owner.ID))
	require.NoError(t, err)
	unowned, err := repo.Create(context.Background(), newTestImage(other.ID))
	require.NoError(t, err)
	nonExistent := uuid.New()

	result, err := repo.FilterOwnedImageIDs(context.Background(), []uuid.UUID{owned.ID, unowned.ID, nonExistent}, owner.ID)

	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{owned.ID}, result)
}

func TestImageRepository_FilterOwnedImageIDs_EmptyInput(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_filterowned_empty")
	require.NoError(t, err)

	repo := NewImageRepository(tx)

	result, err := repo.FilterOwnedImageIDs(context.Background(), nil, user.ID)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestImageRepository_FilterOwnedImageIDs_AllUnowned(t *testing.T) {
	tx := testutil.NewTestTx(t, testDB)

	userRepo := NewUserRepository(tx)
	user, err := userRepo.GetOrCreate(context.Background(), "kp_filterowned_allunowned")
	require.NoError(t, err)
	other, err := userRepo.GetOrCreate(context.Background(), "kp_filterowned_allunowned_other")
	require.NoError(t, err)

	repo := NewImageRepository(tx)
	unowned, err := repo.Create(context.Background(), newTestImage(other.ID))
	require.NoError(t, err)

	result, err := repo.FilterOwnedImageIDs(context.Background(), []uuid.UUID{unowned.ID}, user.ID)

	require.NoError(t, err)
	assert.Empty(t, result)
}
