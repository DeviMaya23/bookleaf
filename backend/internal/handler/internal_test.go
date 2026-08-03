package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

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

func (s *spyInternalShareUsecase) GetPublicFoldersByUser(_ context.Context, _ string) ([]usecase.FolderShareSummary, error) {
	return s.summaries, s.err
}

func (s *spyInternalShareUsecase) GetSharedFolderByFolderID(_ context.Context, _ uuid.UUID) (*usecase.SharedFolder, error) {
	return s.sharedFolder, s.err
}

func (s *spyInternalShareUsecase) CheckFolderPublicStatus(_ context.Context, _ uuid.UUID) (string, error) {
	return s.token, s.err
}

// --- ListPublicFolders ---

func TestInternalHandler_ListPublicFolders_WithFolders(t *testing.T) {
	folderID1 := uuid.New()
	folderID2 := uuid.New()
	spy := &spyInternalShareUsecase{
		summaries: []usecase.FolderShareSummary{
			{FolderID: folderID1, Token: "tok_one", FolderName: "Travel"},
			{FolderID: folderID2, Token: "tok_two", FolderName: "Family"},
		},
	}
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
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
	spy := &spyInternalShareUsecase{summaries: []usecase.FolderShareSummary{}}
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/folders/"+folderID.String()+"/contents", "")
	c.SetPath("/internal/folders/:folder_id/contents")
	c.SetParamNames("folder_id")
	c.SetParamValues(folderID.String())

	err := h.GetFolderContents(c)

	assertHTTPError(t, err, http.StatusNotFound)
}

func TestInternalHandler_GetFolderContents_InvalidUUID(t *testing.T) {
	spy := &spyInternalShareUsecase{}
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/internal/folders/"+folderID.String()+"/status", "")
	c.SetPath("/internal/folders/:folder_id/status")
	c.SetParamNames("folder_id")
	c.SetParamValues(folderID.String())

	err := h.CheckFolderStatus(c)

	assertHTTPError(t, err, http.StatusNotFound)
}

func TestInternalHandler_CheckFolderStatus_InvalidUUID(t *testing.T) {
	spy := &spyInternalShareUsecase{}
	h := NewInternalHandler(spy, nil, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, observability.NewTelemetry(nil, nil, nil))
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
	h := NewInternalHandler(&spyInternalShareUsecase{}, accountSpy, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/internal/accounts/"+userID, "")
	c.SetPath("/internal/accounts/:id")
	c.SetParamNames("id")
	c.SetParamValues(userID)

	err := h.DeleteAccount(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
}
