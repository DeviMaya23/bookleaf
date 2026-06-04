package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

type mockImageUsecase struct {
	downloadURL          string
	imageDetail          *usecase.ImageDetail
	imageItem            *usecase.ImageItem
	listImagesResult     *usecase.ListImagesResult
	listTrashedResult    *usecase.ListTrashedResult
	err                  error
	moveImageFolderErr   error
	lastUpdateParams     usecase.UpdateImageParams
	lastListImagesParams usecase.ListImagesParams
	lastMoveImageID      uuid.UUID
	lastMoveUserID       string
	lastMoveFromFolderID *uuid.UUID
	lastMoveToFolderID   *uuid.UUID
	moveImageFolderCalls int
}

func (m *mockImageUsecase) ListImages(_ context.Context, _ string, params usecase.ListImagesParams) (*usecase.ListImagesResult, error) {
	m.lastListImagesParams = params
	if m.err != nil {
		return nil, m.err
	}
	if m.listImagesResult != nil {
		return m.listImagesResult, nil
	}
	return &usecase.ListImagesResult{}, nil
}

func (m *mockImageUsecase) GetImage(_ context.Context, _ uuid.UUID, _ string) (*usecase.ImageDetail, error) {
	return m.imageDetail, m.err
}

func (m *mockImageUsecase) DownloadImage(_ context.Context, _ uuid.UUID, _ string) (string, error) {
	return m.downloadURL, m.err
}

func (m *mockImageUsecase) UpdateImage(_ context.Context, _ uuid.UUID, _ string, params usecase.UpdateImageParams) (*usecase.ImageItem, error) {
	m.lastUpdateParams = params
	return m.imageItem, m.err
}

func (m *mockImageUsecase) MoveImageFolder(_ context.Context, imageID uuid.UUID, userID string, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) (*usecase.ImageItem, error) {
	m.moveImageFolderCalls++
	m.lastMoveImageID = imageID
	m.lastMoveUserID = userID
	m.lastMoveFromFolderID = fromFolderID
	m.lastMoveToFolderID = toFolderID
	if m.moveImageFolderErr != nil {
		return nil, m.moveImageFolderErr
	}
	return m.imageItem, m.err
}

func (m *mockImageUsecase) UpdateImagePosition(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageUsecase) SoftDelete(_ context.Context, _ uuid.UUID, _ string) error {
	return m.err
}

func (m *mockImageUsecase) ListTrashed(_ context.Context, _ string, _ usecase.ListTrashedParams) (*usecase.ListTrashedResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.listTrashedResult != nil {
		return m.listTrashedResult, nil
	}
	return &usecase.ListTrashedResult{}, nil
}

func (m *mockImageUsecase) Restore(_ context.Context, _ uuid.UUID, _ string) (*usecase.ImageItem, error) {
	return m.imageItem, m.err
}

// --- toImageResponse ---

func TestImageHandler_toImageResponse_FolderIDs(t *testing.T) {
	folderIDOne := uuid.New()
	folderIDTwo := uuid.New()

	t.Run("returns empty folder_ids when ImageFolders is empty", func(t *testing.T) {
		item := usecase.ImageItem{Image: &domain.Image{ID: uuid.New(), Title: "photo"}}
		resp := toImageResponse(item)
		assert.Empty(t, resp.FolderIDs)
	})

	t.Run("returns all folder_ids from ImageFolders", func(t *testing.T) {
		item := usecase.ImageItem{Image: &domain.Image{
			ID:    uuid.New(),
			Title: "photo",
			ImageFolders: []domain.ImageFolder{
				{ImageID: uuid.New(), FolderID: folderIDOne, Position: "1"},
				{ImageID: uuid.New(), FolderID: folderIDTwo, Position: "2"},
			},
		}}
		resp := toImageResponse(item)
		require.Len(t, resp.FolderIDs, 2)
		assert.Equal(t, folderIDOne, resp.FolderIDs[0])
		assert.Equal(t, folderIDTwo, resp.FolderIDs[1])
	})
}

// --- ListImages ---

func TestImageHandler_ListImages_ReturnsList(t *testing.T) {
	tagID := uuid.New()
	uc := &mockImageUsecase{
		listImagesResult: &usecase.ListImagesResult{
			Images: []usecase.ImageItem{
				{Image: &domain.Image{
					ID:    uuid.New(),
					Title: "photo 1",
					Tags:  []domain.Tag{{ID: tagID, Name: "travel"}},
				}},
				{Image: &domain.Image{ID: uuid.New(), Title: "photo 2"}},
			},
		},
	}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/images", "")

	err := h.ListImages(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	images, ok := resp["images"].([]any)
	require.True(t, ok)
	assert.Len(t, images, 2)
	firstImage := images[0].(map[string]any)
	tags := firstImage["tags"].([]any)
	require.Len(t, tags, 1)
	assert.Equal(t, "travel", tags[0].(map[string]any)["name"])
	_, hasCursor := resp["next_cursor"]
	assert.True(t, hasCursor)
}

func TestImageHandler_ListImages_GenericError(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images", "")

	err := h.ListImages(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestImageHandler_ListImages_InvalidCursor(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images?cursor=!!!notvalid!!!", "")

	err := h.ListImages(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_ListImages_EncodesNextCursor(t *testing.T) {
	cursorID := uuid.New()
	uc := &mockImageUsecase{
		listImagesResult: &usecase.ListImagesResult{
			Images: []usecase.ImageItem{{Image: &domain.Image{ID: uuid.New(), Title: "photo"}}},
			NextCursor: &usecase.ImageCursor{
				CreatedAt: time.Now().UTC(),
				ID:        cursorID,
			},
		},
	}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/images", "")

	err := h.ListImages(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["next_cursor"])
	assert.IsType(t, "", resp["next_cursor"])
}

func TestImageHandler_ListImages_UnfiledTrue(t *testing.T) {
	uc := &mockImageUsecase{}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images?unfiled=true", "")

	err := h.ListImages(c)

	require.NoError(t, err)
	assert.True(t, uc.lastListImagesParams.Unfiled)
}

func TestImageHandler_ListImages_UnfiledAbsent(t *testing.T) {
	uc := &mockImageUsecase{}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images", "")

	err := h.ListImages(c)

	require.NoError(t, err)
	assert.False(t, uc.lastListImagesParams.Unfiled)
}

func TestImageHandler_ListImages_InvalidTagID(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images?tag_id=not-a-uuid", "")

	err := h.ListImages(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_ListImages_ValidTagID(t *testing.T) {
	tagID := uuid.New()
	uc := &mockImageUsecase{}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images?tag_id="+tagID.String(), "")

	err := h.ListImages(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastListImagesParams.TagID)
	assert.Equal(t, tagID, *uc.lastListImagesParams.TagID)
}

func TestImageHandler_ListImages_FolderView_IgnoresCursor(t *testing.T) {
	folderID := uuid.New()
	uc := &mockImageUsecase{
		listImagesResult: &usecase.ListImagesResult{
			Images: []usecase.ImageItem{{Image: &domain.Image{ID: uuid.New(), Title: "img"}}},
		},
	}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/images?folder_id="+folderID.String()+"&cursor=somevalue", "")

	err := h.ListImages(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Nil(t, resp["next_cursor"])
}

// --- GetImage ---

func TestImageHandler_GetImage(t *testing.T) {
	imageID := uuid.New()
	tagID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns 200 with image detail",
			uc: &mockImageUsecase{
				imageDetail: &usecase.ImageDetail{
					Image: &domain.Image{
						ID:       imageID,
						Title:    "photo",
						MIMEType: "image/jpeg",
						Tags:     []domain.Tag{{ID: tagID, Name: "travel"}},
					},
					ImageURL: "https://r2.example.com/view",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/images/"+imageID.String(), "")
			c.SetPath("/images/:id")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.GetImage(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, imageID.String(), resp["id"])
			assert.Equal(t, "https://r2.example.com/view", resp["image_url"])
			tags, ok := resp["tags"].([]any)
			require.True(t, ok)
			require.Len(t, tags, 1)
			assert.Equal(t, "travel", tags[0].(map[string]any)["name"])
		})
	}
}

func TestImageHandler_GetImage_InvalidUUID(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images/not-a-uuid", "")
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.GetImage(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- DownloadImage ---

func TestImageHandler_DownloadImage(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with download url",
			uc:         &mockImageUsecase{downloadURL: "https://r2.example.com/download"},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/images/"+imageID.String()+"/download", "")
			c.SetPath("/images/:id/download")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.DownloadImage(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, "https://r2.example.com/download", resp["download_url"])
		})
	}
}

func TestImageHandler_DownloadImage_InvalidUUID(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images/not-a-uuid/download", "")
	c.SetPath("/images/:id/download")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.DownloadImage(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- SoftDelete ---

func TestImageHandler_SoftDelete(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 204 on success",
			uc:         &mockImageUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
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

func TestImageHandler_SoftDelete_InvalidUUID(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/images/not-a-uuid", "")
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.SoftDelete(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- ListTrashed ---

func TestImageHandler_ListTrashed_ReturnsList(t *testing.T) {
	uc := &mockImageUsecase{
		listTrashedResult: &usecase.ListTrashedResult{
			Images: []usecase.ImageItem{
				{Image: &domain.Image{ID: uuid.New(), Title: "deleted photo"}},
			},
		},
	}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/images/trash", "")

	err := h.ListTrashed(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	images, ok := resp["images"].([]any)
	require.True(t, ok)
	assert.Len(t, images, 1)
	_, hasCursor := resp["next_cursor"]
	assert.True(t, hasCursor)
}

func TestImageHandler_ListTrashed_GenericError(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images/trash", "")

	err := h.ListTrashed(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestImageHandler_ListTrashed_InvalidCursor(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/images/trash?cursor=!!!notvalid!!!", "")

	err := h.ListTrashed(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_ListTrashed_EncodesNextCursor(t *testing.T) {
	cursorID := uuid.New()
	uc := &mockImageUsecase{
		listTrashedResult: &usecase.ListTrashedResult{
			Images: []usecase.ImageItem{{Image: &domain.Image{ID: uuid.New(), Title: "deleted photo"}}},
			NextCursor: &usecase.ImageCursor{
				CreatedAt: time.Now().UTC(),
				ID:        cursorID,
			},
		},
	}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
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

func TestImageHandler_Restore(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with restored image",
			uc:         &mockImageUsecase{imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID, Title: "photo"}}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			uc:            &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			uc:            &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
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

func TestImageHandler_Restore_InvalidUUID(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/not-a-uuid/restore", "")
	c.SetPath("/images/:id/restore")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.Restore(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- UpdateImage ---

func TestImageHandler_UpdateImage(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		body          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns 200 with updated image",
			body:       `{"title":"updated title"}`,
			uc:         &mockImageUsecase{imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID, Title: "updated title"}}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			body:          `{"title":"updated title"}`,
			uc:            &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			body:          `{"title":"updated title"}`,
			uc:            &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), tt.body)
			c.SetPath("/images/:id")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.UpdateImage(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, imageID.String(), resp["id"])
			assert.Equal(t, "updated title", resp["title"])
		})
	}
}

func TestImageHandler_UpdateImage_EmptyTitleReturns400(t *testing.T) {
	imageID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), `{"title":""}`)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.UpdateImage(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_UpdateImage_PassesTags(t *testing.T) {
	imageID := uuid.New()
	tagOne, tagTwo := uuid.New(), uuid.New()
	uc := &mockImageUsecase{imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID}}}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	body := fmt.Sprintf(`{"tags":["%s","%s"]}`, tagOne.String(), tagTwo.String())
	c, _ := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), body)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.UpdateImage(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastUpdateParams.Tags)
	assert.Equal(t, []uuid.UUID{tagOne, tagTwo}, *uc.lastUpdateParams.Tags)
}

func TestImageHandler_UpdateImage_ClearsTagsWithEmptyArray(t *testing.T) {
	imageID := uuid.New()
	uc := &mockImageUsecase{imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID}}}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), `{"tags":[]}`)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.UpdateImage(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastUpdateParams.Tags)
	assert.Len(t, *uc.lastUpdateParams.Tags, 0)
}

func TestImageHandler_UpdateImage_PassesFolderIDs(t *testing.T) {
	imageID := uuid.New()
	folderOne, folderTwo := uuid.New(), uuid.New()
	uc := &mockImageUsecase{imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID}}}
	h := NewImageHandler(uc, observability.NewTelemetry(nil, nil, nil))
	body := fmt.Sprintf(`{"folder_ids":["%s","%s"]}`, folderOne.String(), folderTwo.String())
	c, _ := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), body)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.UpdateImage(c)

	require.NoError(t, err)
	require.NotNil(t, uc.lastUpdateParams.FolderIDs)
	assert.Equal(t, []uuid.UUID{folderOne, folderTwo}, *uc.lastUpdateParams.FolderIDs)
}

func TestImageHandler_UpdateImage_InvalidTagID(t *testing.T) {
	imageID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), `{"tags":["not-a-uuid"]}`)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.UpdateImage(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_UpdateImage_InvalidUUID(t *testing.T) {
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/images/not-a-uuid", `{"title":"updated"}`)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.UpdateImage(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_UpdateImage_MalformedJSON(t *testing.T) {
	imageID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/images/"+imageID.String(), `{not valid json}`)
	c.SetPath("/images/:id")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.UpdateImage(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- MoveImageFolder ---

func TestImageHandler_MoveImageFolder(t *testing.T) {
	imageID := uuid.New()
	fromFolderID := uuid.New()
	toFolderID := uuid.New()

	tests := []struct {
		name          string
		body          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns 200 with moved image",
			body: fmt.Sprintf(`{"from_folder_id":"%s","to_folder_id":"%s"}`, fromFolderID.String(), toFolderID.String()),
			uc: &mockImageUsecase{
				imageItem: &usecase.ImageItem{Image: &domain.Image{ID: imageID, Title: "photo"}},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			body:          fmt.Sprintf(`{"from_folder_id":"%s","to_folder_id":"%s"}`, fromFolderID.String(), toFolderID.String()),
			uc:            &mockImageUsecase{moveImageFolderErr: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			body:          fmt.Sprintf(`{"from_folder_id":"%s","to_folder_id":"%s"}`, fromFolderID.String(), toFolderID.String()),
			uc:            &mockImageUsecase{moveImageFolderErr: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/move-folder", tt.body)
			c.SetPath("/images/:id/move-folder")
			c.SetParamNames("id")
			c.SetParamValues(imageID.String())

			err := h.MoveImageFolder(c)

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

func TestImageHandler_MoveImageFolder_MissingFromFolderID(t *testing.T) {
	imageID := uuid.New()
	toFolderID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/move-folder",
		fmt.Sprintf(`{"to_folder_id":"%s"}`, toFolderID.String()))
	c.SetPath("/images/:id/move-folder")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.MoveImageFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_MoveImageFolder_MissingToFolderID(t *testing.T) {
	imageID := uuid.New()
	fromFolderID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/move-folder",
		fmt.Sprintf(`{"from_folder_id":"%s"}`, fromFolderID.String()))
	c.SetPath("/images/:id/move-folder")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.MoveImageFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_MoveImageFolder_InvalidImageUUID(t *testing.T) {
	fromFolderID := uuid.New()
	toFolderID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/not-a-uuid/move-folder",
		fmt.Sprintf(`{"from_folder_id":"%s","to_folder_id":"%s"}`, fromFolderID.String(), toFolderID.String()))
	c.SetPath("/images/:id/move-folder")
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	err := h.MoveImageFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestImageHandler_MoveImageFolder_MalformedJSON(t *testing.T) {
	imageID := uuid.New()
	h := NewImageHandler(&mockImageUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPost, "/images/"+imageID.String()+"/move-folder", `{not valid json}`)
	c.SetPath("/images/:id/move-folder")
	c.SetParamNames("id")
	c.SetParamValues(imageID.String())

	err := h.MoveImageFolder(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- UpdateImagePosition ---

func TestImageHandler_UpdateImagePosition(t *testing.T) {
	imageID := uuid.New()
	folderID := uuid.New()

	tests := []struct {
		name          string
		imageIDParam  string
		body          string
		uc            *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:         "returns 204 on success",
			imageIDParam: imageID.String(),
			body:         `{"folder_id":"` + folderID.String() + `","position":"a1V"}`,
			uc:           &mockImageUsecase{},
			wantStatus:   http.StatusNoContent,
		},
		{
			name:          "returns 400 for empty position",
			imageIDParam:  imageID.String(),
			body:          `{"folder_id":"` + folderID.String() + `","position":""}`,
			uc:            &mockImageUsecase{},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 400 for missing folder_id",
			imageIDParam:  imageID.String(),
			body:          `{"position":"a1V"}`,
			uc:            &mockImageUsecase{},
			wantErrStatus: http.StatusBadRequest,
		},
		{
			name:          "returns 404 when image not found",
			imageIDParam:  imageID.String(),
			body:          `{"folder_id":"` + folderID.String() + `","position":"a1V"}`,
			uc:            &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
		{
			name:          "returns 500 on generic error",
			imageIDParam:  imageID.String(),
			body:          `{"folder_id":"` + folderID.String() + `","position":"a1V"}`,
			uc:            &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
		{
			name:          "returns 400 for invalid image id",
			imageIDParam:  "not-a-uuid",
			body:          `{"folder_id":"` + folderID.String() + `","position":"a1V"}`,
			uc:            &mockImageUsecase{},
			wantErrStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.uc, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodPatch, "/images/"+tt.imageIDParam+"/position", tt.body)
			c.SetParamNames("id")
			c.SetParamValues(tt.imageIDParam)

			err := h.UpdateImagePosition(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
