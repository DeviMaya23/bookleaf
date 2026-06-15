package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	authmw "github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- mock ---

type mockShareUsecase struct {
	token        string
	created      bool
	sharedFolder *usecase.SharedFolder
	err          error
}

func (m *mockShareUsecase) CreateShare(_ context.Context, _ uuid.UUID, _ string) (string, bool, error) {
	return m.token, m.created, m.err
}

func (m *mockShareUsecase) GetShare(_ context.Context, _ uuid.UUID, _ string) (string, error) {
	return m.token, m.err
}

func (m *mockShareUsecase) DeleteShare(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockShareUsecase) GetSharedFolder(_ context.Context, _ string) (*usecase.SharedFolder, error) {
	return m.sharedFolder, m.err
}

// --- CreateShare ---

func TestShareHandler_CreateShare(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockShareUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 201 when a new share is created",
			uc:         &mockShareUsecase{token: "tok_abc123", created: true},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "returns 200 when a share already exists",
			uc:         &mockShareUsecase{token: "tok_abc123", created: false},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when folder not owned",
			uc:            &mockShareUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockShareUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewShareHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/folders/"+folderID.String()+"/share", "")
			c.SetPath("/folders/:id/share")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.CreateShare(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var body shareTokenResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "tok_abc123", body.Token)
		})
	}
}

func TestShareHandler_CreateShare_InvalidUUID(t *testing.T) {
	h := NewShareHandler(&mockShareUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/folders/not-a-uuid/share", "")
	c.SetPath("/folders/:id/share")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.CreateShare(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestShareHandler_CreateShare_MissingAuthContext(t *testing.T) {
	folderID := uuid.New()
	h := NewShareHandler(&mockShareUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/folders/"+folderID.String()+"/share", "")
	c.SetPath("/folders/:id/share")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "")

	err := h.CreateShare(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- GetShare ---

func TestShareHandler_GetShare(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockShareUsecase
		wantErrStatus int
	}{
		{
			name: "returns 200 with token for a shared folder",
			uc:   &mockShareUsecase{token: "tok_abc123"},
		},
		{
			name:          "returns 404 when folder not shared",
			uc:            &mockShareUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockShareUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewShareHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/folders/"+folderID.String()+"/share", "")
			c.SetPath("/folders/:id/share")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.GetShare(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var body shareTokenResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "tok_abc123", body.Token)
		})
	}
}

func TestShareHandler_GetShare_InvalidUUID(t *testing.T) {
	h := NewShareHandler(&mockShareUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/folders/not-a-uuid/share", "")
	c.SetPath("/folders/:id/share")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetShare(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestShareHandler_GetShare_MissingAuthContext(t *testing.T) {
	folderID := uuid.New()
	h := NewShareHandler(&mockShareUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/folders/"+folderID.String()+"/share", "")
	c.SetPath("/folders/:id/share")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "")

	err := h.GetShare(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- DeleteShare ---

func TestShareHandler_DeleteShare(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockShareUsecase
		wantErrStatus int
	}{
		{
			name: "returns 204 on success",
			uc:   &mockShareUsecase{},
		},
		{
			name:          "returns 404 when folder not owned",
			uc:            &mockShareUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockShareUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewShareHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodDelete, "/folders/"+folderID.String()+"/share", "")
			c.SetPath("/folders/:id/share")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.DeleteShare(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

func TestShareHandler_DeleteShare_InvalidUUID(t *testing.T) {
	h := NewShareHandler(&mockShareUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/folders/not-a-uuid/share", "")
	c.SetPath("/folders/:id/share")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.DeleteShare(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestShareHandler_DeleteShare_MissingAuthContext(t *testing.T) {
	folderID := uuid.New()
	h := NewShareHandler(&mockShareUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/folders/"+folderID.String()+"/share", "")
	c.SetPath("/folders/:id/share")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "")

	err := h.DeleteShare(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- GetSharedFolder ---

func TestShareHandler_GetSharedFolder_Success(t *testing.T) {
	notes := "trip board"
	thumbnailURL := "https://r2.example.com/thumb.jpg"
	uc := &mockShareUsecase{
		sharedFolder: &usecase.SharedFolder{
			Name:  "travel",
			Notes: &notes,
			Images: []usecase.SharedImage{
				{Title: "first", ThumbnailURL: &thumbnailURL, FullResURL: "https://r2.example.com/full.jpg"},
				{Title: "second", ThumbnailURL: nil, FullResURL: "https://r2.example.com/full2.jpg"},
			},
		},
	}
	h := NewShareHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/share/tok_abc123", "")
	c.SetPath("/share/:token")
	c.SetParamNames("token")
	c.SetParamValues("tok_abc123")

	err := h.GetSharedFolder(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body sharedFolderResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "travel", body.Folder.Name)
	require.NotNil(t, body.Folder.Notes)
	assert.Equal(t, "trip board", *body.Folder.Notes)
	require.Len(t, body.Images, 2)
	assert.Equal(t, "first", body.Images[0].Title)
	require.NotNil(t, body.Images[0].ThumbnailURL)
	assert.Equal(t, thumbnailURL, *body.Images[0].ThumbnailURL)
	assert.Equal(t, "https://r2.example.com/full.jpg", body.Images[0].FullResURL)
	assert.Equal(t, "second", body.Images[1].Title)
	assert.Nil(t, body.Images[1].ThumbnailURL)
}

func TestShareHandler_GetSharedFolder_UnknownToken(t *testing.T) {
	h := NewShareHandler(&mockShareUsecase{err: gorm.ErrRecordNotFound}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/share/unknown-token", "")
	c.SetPath("/share/:token")
	c.SetParamNames("token")
	c.SetParamValues("unknown-token")

	err := h.GetSharedFolder(c)

	assertHTTPError(t, err, http.StatusNotFound)
}
