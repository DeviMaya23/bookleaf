package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)


// --- mocks ---

type mockImageUsecase struct {
	uploadResult          *usecase.UploadInitResult
	completeResult        *usecase.CompleteUploadResult
	imageDetail           *usecase.ImageDetail
	image                 *domain.Image
	listImagesResult      *usecase.ListImagesResult
	listTrashedResult     *usecase.ListTrashedResult
	err                   error
	lastDescription       *string
	lastUpdateParams      usecase.UpdateImageParams
	lastListImagesParams  usecase.ListImagesParams
}

func (m *mockImageUsecase) InitiateUpload(_ context.Context, _, _, _ string, _ *string, _ *uuid.UUID, description *string) (*usecase.UploadInitResult, error) {
	m.lastDescription = description
	return m.uploadResult, m.err
}

func (m *mockImageUsecase) CompleteUpload(_ context.Context, _ uuid.UUID, _ string) (*usecase.CompleteUploadResult, error) {
	return m.completeResult, m.err
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

func (m *mockImageUsecase) Restore(_ context.Context, _ uuid.UUID, _ string) (*domain.Image, error) {
	return m.image, m.err
}

func (m *mockImageUsecase) UpdateImage(_ context.Context, _ uuid.UUID, _ string, params usecase.UpdateImageParams) (*domain.Image, error) {
	m.lastUpdateParams = params
	return m.image, m.err
}

type mockImageStorageService struct{}

func (m *mockImageStorageService) GeneratePresignedPutURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "", nil
}

func (m *mockImageStorageService) GeneratePresignedGetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}

func (m *mockImageStorageService) GetObject(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockImageStorageService) PutObject(_ context.Context, _ string, _ io.Reader, _ string) error {
	return nil
}

func (m *mockImageStorageService) Ping(_ context.Context) error {
	return nil
}

func (m *mockImageStorageService) CDNUrl(_ string) string {
	return "https://cdn.example.com/thumbnail.jpg"
}

// --- tests ---

func TestImageHandler_InitiateUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		body          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "creates image and returns 201 with upload url",
			body: `{"title":"sunset","mime_type":"image/jpeg","description":"cover"}`,
			mockUC: &mockImageUsecase{
				uploadResult: &usecase.UploadInitResult{
					Image:     &domain.Image{ID: imageID, Title: "sunset"},
					UploadURL: "https://r2.example.com/upload",
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:          "returns 400 for invalid title",
			body:          `{"title":"","mime_type":"image/jpeg"}`,
			mockUC:        &mockImageUsecase{err: usecase.ErrInvalidImageTitle},
			wantErrStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
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
			if tt.name == "creates image and returns 201 with upload url" {
				require.NotNil(t, tt.mockUC.lastDescription)
				assert.Equal(t, "cover", *tt.mockUC.lastDescription)
			}
		})
	}
}

func TestImageHandler_CompleteUpload(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantImageID   string
		wantWarning   string
		wantIsNew     *bool
		wantErrStatus int
	}{
		{
			name: "completes upload and returns 200 response",
			mockUC: &mockImageUsecase{
				completeResult: &usecase.CompleteUploadResult{
					ImageID: imageID,
					FolderSuggestion: &usecase.FolderSuggestion{
						FolderName: "Nature",
						IsNew:      true,
					},
				},
			},
			wantStatus:  http.StatusOK,
			wantImageID: imageID.String(),
			wantIsNew:   func() *bool { v := true; return &v }(),
		},
		{
			name: "completes upload with warning",
			mockUC: &mockImageUsecase{
				completeResult: &usecase.CompleteUploadResult{
					ImageID: imageID,
					Warning: "ai labelling failed",
				},
			},
			wantStatus:  http.StatusOK,
			wantImageID: imageID.String(),
			wantWarning: "ai labelling failed",
		},
		{
			name:          "returns 404 when image not found",
			mockUC:        &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
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
			assert.Equal(t, tt.wantImageID, resp["image_id"])
			if tt.wantWarning != "" {
				assert.Equal(t, tt.wantWarning, resp["warning"])
			} else {
				_, exists := resp["warning"]
				assert.False(t, exists)
			}
			if tt.wantIsNew != nil {
				suggestion, ok := resp["folder_suggestion"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, *tt.wantIsNew, suggestion["is_new"])
			}
		})
	}
}

func TestImageHandler_ListImages(t *testing.T) {
	tests := []struct {
		name          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantLen       int
		wantErrStatus int
	}{
		{
			name: "returns paginated image list",
			mockUC: &mockImageUsecase{
				listImagesResult: &usecase.ListImagesResult{
					Images: []*domain.Image{
						{
							ID:          uuid.New(),
							Title:       "photo 1",
							Description: func() *string { v := "desc"; return &v }(),
							Width:       func() *int { v := 640; return &v }(),
							Height:      func() *int { v := 480; return &v }(),
							FileSize:    func() *int64 { v := int64(1024); return &v }(),
						},
						{ID: uuid.New(), Title: "photo 2"},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:          "returns 500 on usecase error",
			mockUC:        &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/images", "")

			err := h.ListImages(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			images, ok := resp["images"].([]any)
			require.True(t, ok)
			assert.Len(t, images, tt.wantLen)
			_, hasCursor := resp["next_cursor"]
			assert.True(t, hasCursor)
		})
	}
}

func TestImageHandler_ListImages_Pagination(t *testing.T) {
	t.Run("returns 400 for invalid cursor param", func(t *testing.T) {
		h := NewImageHandler(&mockImageUsecase{}, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
		c, _ := newEchoContext(t, http.MethodGet, "/images?cursor=!!!notvalid!!!", "")

		err := h.ListImages(c)

		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("returns paginated envelope with next_cursor on success", func(t *testing.T) {
		cursorID := uuid.New()
		cursorTime := time.Now().UTC()
		mockUC := &mockImageUsecase{
			listImagesResult: &usecase.ListImagesResult{
				Images: []*domain.Image{{ID: uuid.New(), Title: "photo"}},
				NextCursor: &usecase.ImageCursor{
					CreatedAt: cursorTime,
					ID:        cursorID,
				},
			},
		}
		h := NewImageHandler(mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
		c, rec := newEchoContext(t, http.MethodGet, "/images", "")

		err := h.ListImages(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotNil(t, resp["next_cursor"])
		assert.IsType(t, "", resp["next_cursor"])
		images, ok := resp["images"].([]any)
		require.True(t, ok)
		assert.Len(t, images, 1)
	})
}

func TestImageHandler_ListImages_Unfiled(t *testing.T) {
	t.Run("unfiled=true sets Unfiled flag on params", func(t *testing.T) {
		mockUC := &mockImageUsecase{}
		h := NewImageHandler(mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
		c, _ := newEchoContext(t, http.MethodGet, "/images?unfiled=true", "")

		err := h.ListImages(c)

		require.NoError(t, err)
		assert.True(t, mockUC.lastListImagesParams.Unfiled)
	})

	t.Run("unfiled absent leaves Unfiled false", func(t *testing.T) {
		mockUC := &mockImageUsecase{}
		h := NewImageHandler(mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
		c, _ := newEchoContext(t, http.MethodGet, "/images", "")

		err := h.ListImages(c)

		require.NoError(t, err)
		assert.False(t, mockUC.lastListImagesParams.Unfiled)
	})
}

func TestImageHandler_GetImage(t *testing.T) {
	imageID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "returns image detail with signed url",
			mockUC: &mockImageUsecase{
				imageDetail: &usecase.ImageDetail{
					Image: &domain.Image{
						ID:          imageID,
						Title:       "photo",
						Description: func() *string { v := "desc"; return &v }(),
						MIMEType:    "image/jpeg",
						Width:       func() *int { v := 640; return &v }(),
						Height:      func() *int { v := 480; return &v }(),
						FileSize:    func() *int64 { v := int64(1024); return &v }(),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					ImageURL: "https://r2.example.com/view",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			mockUC:        &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
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
			_, hasDescription := resp["description"]
			_, hasWidth := resp["width"]
			_, hasHeight := resp["height"]
			_, hasFileSize := resp["file_size"]
			assert.True(t, hasDescription)
			assert.True(t, hasWidth)
			assert.True(t, hasHeight)
			assert.True(t, hasFileSize)
		})
	}
}

func TestImageHandler_SoftDelete(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "soft deletes image and returns 204",
			mockUC:     &mockImageUsecase{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "returns 404 when image not found",
			mockUC:        &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
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

func TestImageHandler_ListTrashed(t *testing.T) {
	tests := []struct {
		name          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantLen       int
		wantErrStatus int
	}{
		{
			name: "returns paginated trashed images",
			mockUC: &mockImageUsecase{
				listTrashedResult: &usecase.ListTrashedResult{
					Images: []*domain.Image{
						{ID: uuid.New(), Title: "deleted photo"},
					},
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    1,
		},
		{
			name:          "returns 500 on usecase error",
			mockUC:        &mockImageUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
			c, rec := newEchoContext(t, http.MethodGet, "/images/trash", "")

			err := h.ListTrashed(c)

			if tt.wantErrStatus != 0 {
				assertHTTPError(t, err, tt.wantErrStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var resp map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			images, ok := resp["images"].([]any)
			require.True(t, ok)
			assert.Len(t, images, tt.wantLen)
			_, hasCursor := resp["next_cursor"]
			assert.True(t, hasCursor)
		})
	}
}

func TestImageHandler_ListTrashed_Pagination(t *testing.T) {
	t.Run("returns 400 for invalid cursor param", func(t *testing.T) {
		h := NewImageHandler(&mockImageUsecase{}, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
		c, _ := newEchoContext(t, http.MethodGet, "/images/trash?cursor=!!!notvalid!!!", "")

		err := h.ListTrashed(c)

		assertHTTPError(t, err, http.StatusBadRequest)
	})

	t.Run("returns paginated envelope with next_cursor on success", func(t *testing.T) {
		cursorID := uuid.New()
		cursorTime := time.Now().UTC()
		mockUC := &mockImageUsecase{
			listTrashedResult: &usecase.ListTrashedResult{
				Images: []*domain.Image{{ID: uuid.New(), Title: "deleted photo"}},
				NextCursor: &usecase.ImageCursor{
					CreatedAt: cursorTime,
					ID:        cursorID,
				},
			},
		}
		h := NewImageHandler(mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
		c, rec := newEchoContext(t, http.MethodGet, "/images/trash", "")

		err := h.ListTrashed(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.NotNil(t, resp["next_cursor"])
		assert.IsType(t, "", resp["next_cursor"])
		images, ok := resp["images"].([]any)
		require.True(t, ok)
		assert.Len(t, images, 1)
	})
}

func TestImageHandler_Restore(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "restores image and returns 200",
			mockUC:     &mockImageUsecase{image: &domain.Image{ID: imageID, Title: "photo"}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			mockUC:        &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
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

func TestImageHandler_UpdateImage(t *testing.T) {
	imageID := uuid.New()
	title := "updated title"

	tests := []struct {
		name          string
		body          string
		mockUC        *mockImageUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name: "updates image and returns 200 with updated image",
			body: `{"title":"updated title","description":"new desc"}`,
			mockUC: &mockImageUsecase{
				image: &domain.Image{ID: imageID, Title: title},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 404 when image not found",
			body:          `{"title":"updated title"}`,
			mockUC:        &mockImageUsecase{err: gorm.ErrRecordNotFound},
			wantErrStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewImageHandler(tt.mockUC, &mockImageStorageService{}, observability.NewTelemetry(nil, nil, nil))
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
			assert.Equal(t, title, resp["title"])
			if tt.name == "updates image and returns 200 with updated image" {
				require.NotNil(t, tt.mockUC.lastUpdateParams.Description)
				assert.Equal(t, "new desc", *tt.mockUC.lastUpdateParams.Description)
			}
		})
	}
}
