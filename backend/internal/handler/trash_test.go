package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- mock ---

type mockTrashUsecase struct {
	imageItem             *usecase.ImageItem
	listTrashedResult     *usecase.ListTrashedResult
	err                   error
	lastListTrashedParams usecase.ListTrashedParams
	bulkTrashResult       int
	bulkTrashErr          error
	lastBulkTrashUser     string
	lastBulkTrashIDs      []uuid.UUID
}

func (m *mockTrashUsecase) SoftDelete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockTrashUsecase) ListTrashed(_ context.Context, _ string, params usecase.ListTrashedParams) (*usecase.ListTrashedResult, error) {
	m.lastListTrashedParams = params
	if m.err != nil {
		return nil, m.err
	}
	if m.listTrashedResult != nil {
		return m.listTrashedResult, nil
	}
	return &usecase.ListTrashedResult{}, nil
}

func (m *mockTrashUsecase) Restore(_ context.Context, _ uuid.UUID, _ string) (*usecase.ImageItem, error) {
	return m.imageItem, m.err
}

func (m *mockTrashUsecase) DeleteFromTrash(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockTrashUsecase) EmptyTrash(_ context.Context, _ string) error {
	return m.err
}

func (m *mockTrashUsecase) BulkTrash(_ context.Context, userID string, imageIDs []uuid.UUID) (int, error) {
	m.lastBulkTrashUser = userID
	m.lastBulkTrashIDs = imageIDs
	return m.bulkTrashResult, m.bulkTrashErr
}

// --- SoftDelete ---

func TestTrashHandler_SoftDelete(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockTrashUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 204 on success",
			uc:         &mockTrashUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockTrashUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockTrashUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTrashHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodDelete, "/images/"+imageID.String(), "")
			c.SetPath("/images/:id")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.SoftDelete(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestTrashHandler_SoftDelete_InvalidUUID(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/images/not-a-uuid", "")
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.SoftDelete(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- ListTrashed ---

func TestTrashHandler_ListTrashed_ReturnsList(t *testing.T) {
	uc := &mockTrashUsecase{
		listTrashedResult: &usecase.ListTrashedResult{
			Images: []usecase.ImageItem{
				{Image: &domain.Image{ID: uuid.New(), Title: "deleted photo"}},
			},
		},
	}
	h := NewTrashHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/images/trash", "")

	err := h.ListTrashed(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	images, ok := resp["images"].([]any)
	require.True(t, ok)
	assert.Len(t, images, 1)
}

func TestTrashHandler_ListTrashed_ParsesAndForwardsName(t *testing.T) {
	uc := &mockTrashUsecase{}
	h := NewTrashHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images/trash?name=heartopia", "")

	err := h.ListTrashed(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastListTrashedParams.Name)
	assert.Equal(t, "heartopia", *uc.lastListTrashedParams.Name)
}

func TestTrashHandler_ListTrashed_BlankNameTreatedAsAbsent(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "absent", query: "/images/trash"},
		{name: "empty string", query: "/images/trash?name="},
		{name: "whitespace only", query: "/images/trash?name=" + url.QueryEscape("   ")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockTrashUsecase{}
			h := NewTrashHandler(uc, observability.NewTelemetry(nil, nil, nil))
			c, _ := newEchoContext(t, http.MethodGet, tt.query, "")

			err := h.ListTrashed(c)

			require.NoError(t, err)
			assert.Nil(t, uc.lastListTrashedParams.Name)
		})
	}
}

func TestTrashHandler_ListTrashed_GenericError(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images/trash", "")

	err := h.ListTrashed(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestTrashHandler_ListTrashed_EncodesNextCursor(t *testing.T) {
	cursorID := uuid.New()
	uc := &mockTrashUsecase{
		listTrashedResult: &usecase.ListTrashedResult{
			Images: []usecase.ImageItem{{Image: &domain.Image{ID: uuid.New(), Title: "deleted photo"}}},
			NextCursor: &usecase.ImageCursor{
				CreatedAt: time.Now().UTC(),
				ID:        cursorID,
			},
		},
	}
	h := NewTrashHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/images/trash", "")

	err := h.ListTrashed(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["next_cursor"])
	assert.IsType(t, "", resp["next_cursor"])
}

// --- Restore ---

func TestTrashHandler_Restore(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockTrashUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with restored image",
			uc:         &mockTrashUsecase{imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID, Title: "photo"}}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockTrashUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockTrashUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTrashHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/restore", "")
			c.SetPath("/images/:id/restore")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.Restore(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, imageID.String(), resp["id"])
		})
	}
}

func TestTrashHandler_Restore_InvalidUUID(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/not-a-uuid/restore", "")
	c.SetPath("/images/:id/restore")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.Restore(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- DeleteFromTrash ---

func TestTrashHandler_DeleteFromTrash(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockTrashUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 204 on success",
			uc:         &mockTrashUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when image not in trash",
			uc:            &mockTrashUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockTrashUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTrashHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodDelete, "/images/trash/"+imageID.String(), "")
			c.SetPath("/images/trash/:id")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.DeleteFromTrash(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestTrashHandler_DeleteFromTrash_InvalidUUID(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/images/trash/not-a-uuid", "")
	c.SetPath("/images/trash/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.DeleteFromTrash(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- EmptyTrash ---

func TestTrashHandler_EmptyTrash(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/images/trash", "")

	err := h.EmptyTrash(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestTrashHandler_EmptyTrash_Error(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/images/trash", "")

	err := h.EmptyTrash(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- BulkTrash ---

func TestTrashHandler_BulkTrash_ValidRequestReturnsCount(t *testing.T) {
	imageID := uuid.New()
	uc := &mockTrashUsecase{bulkTrashResult: 1}
	h := NewTrashHandler(uc, observability.NewTelemetry(nil, nil, nil))
	body := fmt.Sprintf(`{"image_ids": [%q]}`, imageID)
	c, rec := newEchoContext(t, http.MethodPost, "/images/bulk/trash", body)

	err := h.BulkTrash(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float64(1), resp["succeeded_count"])
	assert.Equal(t, []uuid.UUID{imageID}, uc.lastBulkTrashIDs)
}

func TestTrashHandler_BulkTrash_MalformedImageIDReturns400(t *testing.T) {
	h := NewTrashHandler(&mockTrashUsecase{}, observability.NewTelemetry(nil, nil, nil))
	body := `{"image_ids": ["not-a-uuid"]}`
	c, _ := newEchoContext(t, http.MethodPost, "/images/bulk/trash", body)

	err := h.BulkTrash(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}
