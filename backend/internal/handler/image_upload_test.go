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

type mockUploadUsecase struct {
	uploadResult       *usecase.UploadInitResult
	completeResult     *usecase.CompleteUploadResult
	err                error
	lastDescription    *string
	lastWidth          *int
	lastHeight         *int
	lastFileSize       *int64
	backfillEnqueued   int
	backfillErr        error
	lastBackfillUserID string
}

func (m *mockUploadUsecase) InitiateUpload(_ context.Context, _, _, _ string, _ *string, _ *uuid.UUID, description *string) (*usecase.UploadInitResult, error) {
	m.lastDescription = description
	return m.uploadResult, m.err
}

func (m *mockUploadUsecase) CompleteUpload(_ context.Context, _ uuid.UUID, _ string, width, height *int, fileSize *int64, _ *string) (*usecase.CompleteUploadResult, error) {
	m.lastWidth = width
	m.lastHeight = height
	m.lastFileSize = fileSize
	return m.completeResult, m.err
}

func (m *mockUploadUsecase) BackfillVisionLabels(_ context.Context, userID string) (int, error) {
	m.lastBackfillUserID = userID
	return m.backfillEnqueued, m.backfillErr
}

// --- InitiateUpload ---

func TestUploadHandler_InitiateUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		body          string
		uc            *mockUploadUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns 201 with upload body",
			body: `{"title":"sunset","mime_type":"image/jpeg"}`,
			uc: &mockUploadUsecase{
				uploadResult: &usecase.UploadInitResult{
					ID:        imageID,
					UploadURL: "https://r2.example.com/upload",
					R2Path:    "users/kp_abc123/images/" + imageID.String() + ".jpg",
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:          "returns 400 for invalid title",
			body:          `{"title":"","mime_type":"image/jpeg"}`,
			uc:            &mockUploadUsecase{err: usecase.ErrInvalidImageTitle},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 400 for invalid mime type",
			body:          `{"title":"sunset","mime_type":""}`,
			uc:            &mockUploadUsecase{err: usecase.ErrInvalidMIMEType},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 500 on generic error",
			body:          `{"title":"sunset","mime_type":"image/jpeg"}`,
			uc:            &mockUploadUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUploadHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/images", tt.body)

			err := h.InitiateUpload(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, imageID.String(), resp["id"])
			assert.Equal(t, "https://r2.example.com/upload", resp["upload_url"])
			assert.Equal(t, "users/kp_abc123/images/"+imageID.String()+".jpg", resp["r2_path"])
		})
	}
}

func TestUploadHandler_InitiateUpload_MalformedJSON(t *testing.T) {
	h := NewUploadHandler(&mockUploadUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images", `{not valid json}`)

	err := h.InitiateUpload(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestUploadHandler_InitiateUpload_PassesDescription(t *testing.T) {
	uc := &mockUploadUsecase{uploadResult: &usecase.UploadInitResult{ID: uuid.New()}}
	h := NewUploadHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images", `{"title":"sunset","mime_type":"image/jpeg","description":"golden hour"}`)

	err := h.InitiateUpload(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastDescription)
	assert.Equal(t, "golden hour", *uc.lastDescription)
}

// --- CompleteUpload ---

func TestUploadHandler_CompleteUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockUploadUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with image_id",
			uc:         &mockUploadUsecase{completeResult: &usecase.CompleteUploadResult{ImageID: imageID}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockUploadUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockUploadUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUploadHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/complete", "")
			c.SetPath("/images/:id/complete")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.CompleteUpload(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, imageID.String(), resp["image_id"])
		})
	}
}

func TestUploadHandler_CompleteUpload_ForwardsRequestBodyToUsecase(t *testing.T) {
	imageID := uuid.New()
	uc := &mockUploadUsecase{completeResult: &usecase.CompleteUploadResult{ImageID: imageID}}
	h := NewUploadHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/complete", `{"width":1920,"height":1080,"file_size":245760}`)
	c.SetPath("/images/:id/complete")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.CompleteUpload(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastWidth)
	require.NotNil(t, uc.lastHeight)
	require.NotNil(t, uc.lastFileSize)
	assert.Equal(t, 1920, *uc.lastWidth)
	assert.Equal(t, 1080, *uc.lastHeight)
	assert.EqualValues(t, 245760, *uc.lastFileSize)
}

func TestUploadHandler_CompleteUpload_MissingBodyTreatsValuesAsAbsent(t *testing.T) {
	imageID := uuid.New()
	uc := &mockUploadUsecase{completeResult: &usecase.CompleteUploadResult{ImageID: imageID}}
	h := NewUploadHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/complete", "")
	c.SetPath("/images/:id/complete")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.CompleteUpload(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, uc.lastWidth)
	assert.Nil(t, uc.lastHeight)
	assert.Nil(t, uc.lastFileSize)
}

func TestUploadHandler_CompleteUpload_InvalidUUID(t *testing.T) {
	h := NewUploadHandler(&mockUploadUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/not-a-uuid/complete", "")
	c.SetPath("/images/:id/complete")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.CompleteUpload(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- BackfillVision ---

func TestUploadHandler_BackfillVision_Success(t *testing.T) {
	uc := &mockUploadUsecase{backfillEnqueued: 5}
	h := NewUploadHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodPost, "/me/vision/backfill", "")

	err := h.BackfillVision(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.JSONEq(t, `{"enqueued":5}`, rec.Body.String())
	assert.Equal(t, "kp_abc123", uc.lastBackfillUserID)
}

func TestUploadHandler_BackfillVision_UsecaseError(t *testing.T) {
	uc := &mockUploadUsecase{backfillErr: errors.New("db error")}
	h := NewUploadHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/me/vision/backfill", "")

	err := h.BackfillVision(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestUploadHandler_BackfillVision_MissingAuthContext(t *testing.T) {
	h := NewUploadHandler(&mockUploadUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/me/vision/backfill", "")
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "")

	err := h.BackfillVision(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}
