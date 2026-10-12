package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- spies ---

type spyInternalAccountUsecase struct {
	calledWithUserID string
	err              error
}

func (s *spyInternalAccountUsecase) MarkForDeletion(_ context.Context, userID string) error {
	s.calledWithUserID = userID
	return s.err
}

type spyInternalShareUsecase struct {
	summaries    []usecase.FolderShareSummary
	sharedFolder *usecase.SharedFolder
	token        string
	err          error
}

func (s *spyInternalShareUsecase) GetPublicFoldersByUser(_ context.Context, _ uuid.UUID) ([]usecase.FolderShareSummary, error) {
	return s.summaries, s.err
}

func (s *spyInternalShareUsecase) GetSharedFolderByFolderID(_ context.Context, _ uuid.UUID) (*usecase.SharedFolder, error) {
	return s.sharedFolder, s.err
}

func (s *spyInternalShareUsecase) CheckFolderPublicStatus(_ context.Context, _ uuid.UUID) (string, error) {
	return s.token, s.err
}

type spyUserResolver struct {
	user *domain.User
	err  error
}

func (s *spyUserResolver) GetByIDPSubject(_ context.Context, _ string) (*domain.User, error) {
	return s.user, s.err
}

type spyInternalFolderImagesUsecase struct {
	thumbnails []usecase.FolderImageThumbnail
	err        error
}

func (s *spyInternalFolderImagesUsecase) ListFolderImages(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]usecase.FolderImageThumbnail, error) {
	return s.thumbnails, s.err
}

// --- ListPublicFolders ---

func TestInternalHandler_ListPublicFolders_WithFolders(t *testing.T) {
	folderID1 := uuid.New()
	folderID2 := uuid.New()
	resolvedUser := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	resolver := &spyUserResolver{user: resolvedUser}
	shareSpy := &spyInternalShareUsecase{
		summaries: []usecase.FolderShareSummary{
			{FolderID: folderID1, Token: "tok_one", FolderName: "Travel"},
			{FolderID: folderID2, Token: "tok_two", FolderName: "Family"},
		},
	}
	h := NewInternalHandler(shareSpy, nil, resolver, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/users/kp_abc123/public-folders", "")
	c.SetPath("/internal/users/:user_id/public-folders")
	c.SetParamNames("user_id")
	c.SetParamValues("kp_abc123")

	err := h.ListPublicFolders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body listPublicFoldersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.FolderList, 2)
	assert.Equal(t, folderID1.String(), body.FolderList[0].FolderID)
	assert.Equal(t, "tok_one", body.FolderList[0].Token)
	assert.Equal(t, "Travel", body.FolderList[0].FolderName)
	assert.Equal(t, folderID2.String(), body.FolderList[1].FolderID)
	assert.Equal(t, "tok_two", body.FolderList[1].Token)
	assert.Equal(t, "Family", body.FolderList[1].FolderName)
}

func TestInternalHandler_ListPublicFolders_EmptyList(t *testing.T) {
	resolvedUser := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	resolver := &spyUserResolver{user: resolvedUser}
	shareSpy := &spyInternalShareUsecase{summaries: []usecase.FolderShareSummary{}}
	h := NewInternalHandler(shareSpy, nil, resolver, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/users/kp_abc123/public-folders", "")
	c.SetPath("/internal/users/:user_id/public-folders")
	c.SetParamNames("user_id")
	c.SetParamValues("kp_abc123")

	err := h.ListPublicFolders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body listPublicFoldersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Empty(t, body.FolderList)
}

func TestInternalHandler_ListPublicFolders_UnknownIDPSubject(t *testing.T) {
	resolver := &spyUserResolver{err: usecase.ErrUserNotFound}
	h := NewInternalHandler(&spyInternalShareUsecase{}, nil, resolver, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/users/kp_unknown/public-folders", "")
	c.SetPath("/internal/users/:user_id/public-folders")
	c.SetParamNames("user_id")
	c.SetParamValues("kp_unknown")

	err := h.ListPublicFolders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body listPublicFoldersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Empty(t, body.FolderList)
}

// --- GetFolderContents ---

func TestInternalHandler_GetFolderContents_Success(t *testing.T) {
	folderID := uuid.New()
	notes := "trip board"
	thumbnailURL := "https://r2.example.com/thumb.jpg"
	width, height := 800, 600
	spy := &spyInternalShareUsecase{
		sharedFolder: &usecase.SharedFolder{
			Name:  "travel",
			Notes: &notes,
			Images: []usecase.SharedImage{
				{Title: "first", ThumbnailURL: &thumbnailURL, FullResURL: "https://r2.example.com/full.jpg", DownloadURL: "https://r2.example.com/full.jpg?dl", Width: &width, Height: &height},
			},
		},
	}
	h := NewInternalHandler(spy, nil, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/folders/"+folderID.String()+"/contents", "")
	c.SetPath("/internal/folders/:folder_id/contents")
	c.SetParamNames("folder_id")
	c.SetParamValues(folderID.String())

	err := h.GetFolderContents(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body sharedFolderResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "travel", body.Folder.Name)
	require.NotNil(t, body.Folder.Notes)
	assert.Equal(t, "trip board", *body.Folder.Notes)
	require.Len(t, body.Images, 1)
	assert.Equal(t, "first", body.Images[0].Title)
}

func TestInternalHandler_GetFolderContents_NotFound(t *testing.T) {
	folderID := uuid.New()
	spy := &spyInternalShareUsecase{err: gorm.ErrRecordNotFound}
	h := NewInternalHandler(spy, nil, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/folders/"+folderID.String()+"/contents", "")
	c.SetPath("/internal/folders/:folder_id/contents")
	c.SetParamNames("folder_id")
	c.SetParamValues(folderID.String())

	err := h.GetFolderContents(c)

	assertHTTPError(t, err, http.StatusNotFound)
}

func TestInternalHandler_GetFolderContents_InvalidUUID(t *testing.T) {
	spy := &spyInternalShareUsecase{}
	h := NewInternalHandler(spy, nil, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/folders/not-a-uuid/contents", "")
	c.SetPath("/internal/folders/:folder_id/contents")
	c.SetParamNames("folder_id")
	c.SetParamValues("not-a-uuid")

	err := h.GetFolderContents(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- CheckFolderStatus ---

func TestInternalHandler_CheckFolderStatus_Public(t *testing.T) {
	folderID := uuid.New()
	spy := &spyInternalShareUsecase{token: "tok_abc123"}
	h := NewInternalHandler(spy, nil, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/folders/"+folderID.String()+"/status", "")
	c.SetPath("/internal/folders/:folder_id/status")
	c.SetParamNames("folder_id")
	c.SetParamValues(folderID.String())

	err := h.CheckFolderStatus(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body shareTokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "tok_abc123", body.Token)
}

func TestInternalHandler_CheckFolderStatus_NotFound(t *testing.T) {
	folderID := uuid.New()
	spy := &spyInternalShareUsecase{err: gorm.ErrRecordNotFound}
	h := NewInternalHandler(spy, nil, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/folders/"+folderID.String()+"/status", "")
	c.SetPath("/internal/folders/:folder_id/status")
	c.SetParamNames("folder_id")
	c.SetParamValues(folderID.String())

	err := h.CheckFolderStatus(c)

	assertHTTPError(t, err, http.StatusNotFound)
}

func TestInternalHandler_CheckFolderStatus_InvalidUUID(t *testing.T) {
	spy := &spyInternalShareUsecase{}
	h := NewInternalHandler(spy, nil, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/folders/not-a-uuid/status", "")
	c.SetPath("/internal/folders/:folder_id/status")
	c.SetParamNames("folder_id")
	c.SetParamValues("not-a-uuid")

	err := h.CheckFolderStatus(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- DeleteAccount ---

func TestInternalHandler_DeleteAccount_CallsMarkForDeletionAndReturns202(t *testing.T) {
	userID := "kp_abc123"
	accountSpy := &spyInternalAccountUsecase{}
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/internal/accounts/"+userID, "")
	c.SetPath("/internal/accounts/:id")
	c.SetParamNames("id")
	c.SetParamValues(userID)

	err := h.DeleteAccount(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Equal(t, userID, accountSpy.calledWithUserID)
}

func TestInternalHandler_DeleteAccount_ReturnsAcceptedForPendingDeletionUser(t *testing.T) {
	userID := "kp_pending"
	accountSpy := &spyInternalAccountUsecase{}
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/internal/accounts/"+userID, "")
	c.SetPath("/internal/accounts/:id")
	c.SetParamNames("id")
	c.SetParamValues(userID)

	err := h.DeleteAccount(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func TestInternalHandler_DeleteAccount_ReturnsAcceptedForPurgedUser(t *testing.T) {
	userID := "kp_purged"
	accountSpy := &spyInternalAccountUsecase{}
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/internal/accounts/"+userID, "")
	c.SetPath("/internal/accounts/:id")
	c.SetParamNames("id")
	c.SetParamValues(userID)

	err := h.DeleteAccount(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}

func TestInternalHandler_DeleteAccount_ReturnsAcceptedForUnprovisionedUser(t *testing.T) {
	userID := "kp_unknown"
	accountSpy := &spyInternalAccountUsecase{}
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, nil, nil, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/internal/accounts/"+userID, "")
	c.SetPath("/internal/accounts/:id")
	c.SetParamNames("id")
	c.SetParamValues(userID)

	err := h.DeleteAccount(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}

// --- ListFolderImages ---

func TestInternalHandler_ListFolderImages_ReturnsThumbnails(t *testing.T) {
	resolvedUser := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	folderID := uuid.New()
	imageID := uuid.New()
	spy := &spyInternalFolderImagesUsecase{
		thumbnails: []usecase.FolderImageThumbnail{
			{ImageID: imageID, ThumbnailURL: "https://cdn.example.com/thumb.jpg"},
		},
	}
	h := NewInternalHandler(&spyInternalShareUsecase{}, nil, &spyUserResolver{user: resolvedUser}, spy, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/users/kp_abc123/folders/"+folderID.String()+"/images", "")
	c.SetPath("/internal/users/:user_id/folders/:folder_id/images")
	c.SetParamNames("user_id", "folder_id")
	c.SetParamValues("kp_abc123", folderID.String())

	err := h.ListFolderImages(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body listFolderImagesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Images, 1)
	assert.Equal(t, imageID.String(), body.Images[0].ImageID)
	assert.Equal(t, "https://cdn.example.com/thumb.jpg", body.Images[0].ThumbnailURL)
}

func TestInternalHandler_ListFolderImages_EmptyFolderReturns200WithEmptyArray(t *testing.T) {
	resolvedUser := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	folderID := uuid.New()
	spy := &spyInternalFolderImagesUsecase{thumbnails: []usecase.FolderImageThumbnail{}}
	h := NewInternalHandler(&spyInternalShareUsecase{}, nil, &spyUserResolver{user: resolvedUser}, spy, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/internal/users/kp_abc123/folders/"+folderID.String()+"/images", "")
	c.SetPath("/internal/users/:user_id/folders/:folder_id/images")
	c.SetParamNames("user_id", "folder_id")
	c.SetParamValues("kp_abc123", folderID.String())

	err := h.ListFolderImages(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body listFolderImagesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Empty(t, body.Images)
}

func TestInternalHandler_ListFolderImages_FolderNotFoundReturns404(t *testing.T) {
	resolvedUser := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	folderID := uuid.New()
	spy := &spyInternalFolderImagesUsecase{err: gorm.ErrRecordNotFound}
	h := NewInternalHandler(&spyInternalShareUsecase{}, nil, &spyUserResolver{user: resolvedUser}, spy, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/users/kp_abc123/folders/"+folderID.String()+"/images", "")
	c.SetPath("/internal/users/:user_id/folders/:folder_id/images")
	c.SetParamNames("user_id", "folder_id")
	c.SetParamValues("kp_abc123", folderID.String())

	err := h.ListFolderImages(c)

	assertHTTPError(t, err, http.StatusNotFound)
}

func TestInternalHandler_ListFolderImages_UnknownUserReturns404(t *testing.T) {
	folderID := uuid.New()
	spy := &spyInternalFolderImagesUsecase{}
	h := NewInternalHandler(&spyInternalShareUsecase{}, nil, &spyUserResolver{err: usecase.ErrUserNotFound}, spy, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/users/kp_unknown/folders/"+folderID.String()+"/images", "")
	c.SetPath("/internal/users/:user_id/folders/:folder_id/images")
	c.SetParamNames("user_id", "folder_id")
	c.SetParamValues("kp_unknown", folderID.String())

	err := h.ListFolderImages(c)

	assertHTTPError(t, err, http.StatusNotFound)
}

func TestInternalHandler_ListFolderImages_InvalidFolderUUIDReturns400(t *testing.T) {
	resolvedUser := &domain.User{ID: uuid.New(), IDPSubject: "kp_abc123"}
	spy := &spyInternalFolderImagesUsecase{}
	h := NewInternalHandler(&spyInternalShareUsecase{}, nil, &spyUserResolver{user: resolvedUser}, spy, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/users/kp_abc123/folders/not-a-uuid/images", "")
	c.SetPath("/internal/users/:user_id/folders/:folder_id/images")
	c.SetParamNames("user_id", "folder_id")
	c.SetParamValues("kp_abc123", "not-a-uuid")

	err := h.ListFolderImages(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}
