package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	authmw "github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockFolderUsecase struct {
	folder           *domain.Folder
	detail           *usecase.FolderDetail
	folders          []*domain.Folder
	err              error
	lastUpdateParams usecase.UpdateFolderParams
	exportBytes      []byte
	exportErr        error
}

func (m *mockFolderUsecase) Create(_ context.Context, _, _ string, _ *uuid.UUID, _ *string) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderUsecase) List(_ context.Context, _ string) ([]*domain.Folder, error) {
	return m.folders, m.err
}

func (m *mockFolderUsecase) GetByID(_ context.Context, _ uuid.UUID, _ string) (*usecase.FolderDetail, error) {
	return m.detail, m.err
}

func (m *mockFolderUsecase) Update(_ context.Context, _ uuid.UUID, _ string, params usecase.UpdateFolderParams) (*domain.Folder, error) {
	m.lastUpdateParams = params
	return m.folder, m.err
}

func (m *mockFolderUsecase) Delete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockFolderUsecase) ExportFolder(_ context.Context, _ uuid.UUID, _ string, w io.Writer) error {
	if m.exportErr != nil {
		return m.exportErr
	}
	_, err := w.Write(m.exportBytes)
	return err
}

func newEchoContext(t *testing.T, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	var bodyReader *bytes.Reader
	if body != "" {
		bodyReader = bytes.NewReader([]byte(body))
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")
	return c, rec
}

func assertHTTPError(t *testing.T, err error, wantStatus int) {
	t.Helper()
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok, "expected *echo.HTTPError, got %T", err)
	assert.Equal(t, wantStatus, httpErr.Code)
}

// --- CreateFolder ---

func TestFolderHandler_CreateFolder(t *testing.T) {
	folderID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		body          string
		uc            *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns 201 with folder body",
			body: `{"name":"travel","description":"trip board"}`,
			uc: &mockFolderUsecase{
				folder: &domain.Folder{
					ID:          folderID,
					Name:        "travel",
					Description: func() *string { v := "trip board"; return &v }(),
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:          "returns 400 for invalid name",
			body:          `{"name":""}`,
			uc:            &mockFolderUsecase{err: usecase.ErrInvalidFolderName},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 500 on generic error",
			body:          `{"name":"travel"}`,
			uc:            &mockFolderUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/folders", tt.body)

			err := h.CreateFolder(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, folderID.String(), resp["id"])
			assert.Equal(t, "travel", resp["name"])
			_, hasDescription := resp["description"]
			assert.True(t, hasDescription)
		})
	}
}

func TestFolderHandler_CreateFolder_MalformedJSON(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/folders", `{not valid json}`)

	err := h.CreateFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- ListFolders ---

func TestFolderHandler_ListFolders_ReturnsList(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{
		folders: []*domain.Folder{
			{ID: uuid.New(), Name: "travel", Description: func() *string { v := "trip"; return &v }()},
			{ID: uuid.New(), Name: "design"},
		},
	}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/folders", "")

	err := h.ListFolders(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
	_, hasDescription := resp[0]["description"]
	assert.True(t, hasDescription)
}

func TestFolderHandler_ListFolders_GenericError(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/folders", "")

	err := h.ListFolders(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- GetFolder ---

func TestFolderHandler_GetFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns 200 with folder detail",
			uc: &mockFolderUsecase{
				detail: &usecase.FolderDetail{
					Folder:     &domain.Folder{ID: folderID, Name: "travel", Description: func() *string { v := "trip"; return &v }()},
					ImageCount: 3,
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when folder not found",
			uc:            &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockFolderUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/folders/"+folderID.String(), "")
			c.SetPath("/folders/:id")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.GetFolder(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, folderID.String(), resp["id"])
			assert.EqualValues(t, 3, resp["image_count"])
			_, hasDescription := resp["description"]
			assert.True(t, hasDescription)
		})
	}
}

func TestFolderHandler_GetFolder_InvalidUUID(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/folders/not-a-uuid", "")
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- UpdateFolder ---

func TestFolderHandler_UpdateFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		body          string
		uc            *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with updated folder",
			body:       `{"name":"updated","description":"new desc"}`,
			uc:         &mockFolderUsecase{folder: &domain.Folder{ID: folderID, Name: "updated", Description: func() *string { v := "new desc"; return &v }()}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 400 for invalid name",
			body:          `{"name":""}`,
			uc:            &mockFolderUsecase{err: usecase.ErrInvalidFolderName},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 404 when folder not found",
			body:          `{"name":"updated"}`,
			uc:            &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			body:          `{"name":"updated"}`,
			uc:            &mockFolderUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPatch, "/folders/"+folderID.String(), tt.body)
			c.SetPath("/folders/:id")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.UpdateFolder(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, folderID.String(), resp["id"])
			assert.Equal(t, "updated", resp["name"])
			_, hasDescription := resp["description"]
			assert.True(t, hasDescription)
		})
	}
}

func TestFolderHandler_UpdateFolder_AbsentFieldsLeaveParamsNil(t *testing.T) {
	folderID := uuid.New()
	uc := &mockFolderUsecase{folder: &domain.Folder{ID: folderID, Name: "updated"}}
	h := NewFolderHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/folders/"+folderID.String(), `{"name":"updated"}`)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())

	err := h.UpdateFolder(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastUpdateParams.Name)
	assert.Equal(t, "updated", *uc.lastUpdateParams.Name)
	assert.Nil(t, uc.lastUpdateParams.ParentID)
	assert.Nil(t, uc.lastUpdateParams.Description)
}

func TestFolderHandler_UpdateFolder_ExplicitNullClearsParentAndDescription(t *testing.T) {
	folderID := uuid.New()
	uc := &mockFolderUsecase{folder: &domain.Folder{ID: folderID}}
	h := NewFolderHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/folders/"+folderID.String(), `{"parent_id":null,"description":null}`)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())

	err := h.UpdateFolder(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastUpdateParams.ParentID)
	assert.Nil(t, *uc.lastUpdateParams.ParentID)
	require.NotNil(t, uc.lastUpdateParams.Description)
	assert.Nil(t, *uc.lastUpdateParams.Description)
}

func TestFolderHandler_UpdateFolder_ProvidedValuesSetParentAndDescription(t *testing.T) {
	folderID := uuid.New()
	parentID := uuid.New()
	uc := &mockFolderUsecase{folder: &domain.Folder{ID: folderID}}
	h := NewFolderHandler(uc, observability.NewTelemetry(nil, nil, nil))
	body := `{"parent_id":"` + parentID.String() + `","description":"new desc"}`
	c, _ := newEchoContext(t, http.MethodPatch, "/folders/"+folderID.String(), body)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())

	err := h.UpdateFolder(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastUpdateParams.ParentID)
	require.NotNil(t, *uc.lastUpdateParams.ParentID)
	assert.Equal(t, parentID, **uc.lastUpdateParams.ParentID)
	require.NotNil(t, uc.lastUpdateParams.Description)
	require.NotNil(t, *uc.lastUpdateParams.Description)
	assert.Equal(t, "new desc", **uc.lastUpdateParams.Description)
}

func TestFolderHandler_UpdateFolder_InvalidUUID(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/folders/not-a-uuid", `{"name":"updated"}`)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.UpdateFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestFolderHandler_UpdateFolder_MalformedJSON(t *testing.T) {
	folderID := uuid.New()
	h := NewFolderHandler(&mockFolderUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/folders/"+folderID.String(), `{not valid json}`)
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues(folderID.String())

	err := h.UpdateFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- DeleteFolder ---

func TestFolderHandler_DeleteFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 204 on success",
			uc:         &mockFolderUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when folder not found",
			uc:            &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockFolderUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodDelete, "/folders/"+folderID.String(), "")
			c.SetPath("/folders/:id")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.DeleteFolder(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestFolderHandler_DeleteFolder_InvalidUUID(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/folders/not-a-uuid", "")
	c.SetPath("/folders/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.DeleteFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- ExportFolder ---

func TestFolderHandler_ExportFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockFolderUsecase
		wantErrStatus int
		wantFilename  string
	}{
		{
			name: "returns 200 with zip headers",
			uc: &mockFolderUsecase{
				detail:      &usecase.FolderDetail{Folder: &domain.Folder{ID: folderID, Name: "My Folder"}},
				exportBytes: []byte("zip-bytes"),
			},
			wantFilename: `attachment; filename="My Folder.zip"`,
		},
		{
			name: "sanitizes folder name with invalid filename characters",
			uc: &mockFolderUsecase{
				detail: &usecase.FolderDetail{Folder: &domain.Folder{ID: folderID, Name: "Trip / 2024"}},
			},
			wantFilename: `attachment; filename="Trip  2024.zip"`,
		},
		{
			name: "falls back to export.zip when name sanitizes to empty",
			uc: &mockFolderUsecase{
				detail: &usecase.FolderDetail{Folder: &domain.Folder{ID: folderID, Name: "///"}},
			},
			wantFilename: `attachment; filename="export.zip"`,
		},
		{
			name:          "returns 404 when folder not found",
			uc:            &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockFolderUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/folders/"+folderID.String()+"/export", "")
			c.SetPath("/folders/:id/export")
			c.SetParamNames("id")
			c.SetParamValues(folderID.String())

			err := h.ExportFolder(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/zip", rec.Header().Get("Content-Type"))
			assert.Equal(t, tt.wantFilename, rec.Header().Get("Content-Disposition"))
		})
	}
}

func TestFolderHandler_ExportFolder_InvalidUUID(t *testing.T) {
	h := NewFolderHandler(&mockFolderUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/folders/not-a-uuid/export", "")
	c.SetPath("/folders/:id/export")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.ExportFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}
