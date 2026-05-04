package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	authmw "github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockFolderUsecase struct {
	folder  *domain.Folder
	folders []*domain.Folder
	err     error
}

func (m *mockFolderUsecase) Create(_ context.Context, _, _ string, _ *uuid.UUID) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderUsecase) List(_ context.Context, _ string) ([]*domain.Folder, error) {
	return m.folders, m.err
}

func (m *mockFolderUsecase) GetByID(_ context.Context, _ uuid.UUID, _ string) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderUsecase) Update(_ context.Context, _ uuid.UUID, _, _ string, _ *uuid.UUID) (*domain.Folder, error) {
	return m.folder, m.err
}

func (m *mockFolderUsecase) Delete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
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

func TestFolderHandler_CreateFolder(t *testing.T) {
	folderID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		body          string
		mockUC        *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "creates folder and returns 201",
			body: `{"name":"travel"}`,
			mockUC: &mockFolderUsecase{
				folder: &domain.Folder{ID: folderID, Name: "travel", CreatedAt: now, UpdatedAt: now},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:          "returns 400 on invalid name",
			body:          `{"name":""}`,
			mockUC:        &mockFolderUsecase{err: usecase.ErrInvalidFolderName},
			wantErrStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
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
		})
	}
}

func TestFolderHandler_ListFolders(t *testing.T) {
	tests := []struct {
		name          string
		mockUC        *mockFolderUsecase
		wantStatus    int
		wantLen       int
		wantErrStatus int
	}{
		{
			name: "returns folder list",
			mockUC: &mockFolderUsecase{
				folders: []*domain.Folder{
					{ID: uuid.New(), Name: "travel"},
					{ID: uuid.New(), Name: "design"},
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:          "returns 500 on usecase error",
			mockUC:        &mockFolderUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/folders", "")

			err := h.ListFolders(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp []map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Len(t, resp, tt.wantLen)
		})
	}
}

func TestFolderHandler_GetFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		mockUC        *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns folder by id",
			mockUC:     &mockFolderUsecase{folder: &domain.Folder{ID: folderID, Name: "travel"}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when folder not found",
			mockUC:        &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
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
		})
	}
}

func TestFolderHandler_UpdateFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		body          string
		mockUC        *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "updates folder and returns 200",
			body:       `{"name":"updated"}`,
			mockUC:     &mockFolderUsecase{folder: &domain.Folder{ID: folderID, Name: "updated"}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 400 on invalid name",
			body:          `{"name":""}`,
			mockUC:        &mockFolderUsecase{err: usecase.ErrInvalidFolderName},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 404 when folder not found",
			body:          `{"name":"updated"}`,
			mockUC:        &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPut, "/folders/"+folderID.String(), tt.body)
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
		})
	}
}

func TestFolderHandler_DeleteFolder(t *testing.T) {
	folderID := uuid.New()

	tests := []struct {
		name          string
		mockUC        *mockFolderUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "deletes folder and returns 204",
			mockUC:     &mockFolderUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when folder not found",
			mockUC:        &mockFolderUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFolderHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
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
