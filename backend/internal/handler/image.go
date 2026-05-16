package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type ImageHandler struct {
	imageUsecase usecase.ImageUsecase
	store        storage.StorageService
	tel          *observability.Telemetry
}

type updateImageRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	FolderID    json.RawMessage `json:"folder_id"`
}

type initiateImageUploadRequest struct {
	Title       string     `json:"title"`
	MIMEType    string     `json:"mime_type"`
	SourceURL   *string    `json:"source_url"`
	FolderID    *uuid.UUID `json:"folder_id"`
	Description *string    `json:"description"`
}

type initiateImageUploadResponse struct {
	ID        uuid.UUID `json:"id"`
	UploadURL string    `json:"upload_url"`
	R2Path    string    `json:"r2_path"`
}

type imageResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	MIMEType     string     `json:"mime_type"`
	SourceURL    *string    `json:"source_url"`
	FolderID     *uuid.UUID `json:"folder_id"`
	ThumbnailURL *string    `json:"thumbnail_url"`
	Width        *int       `json:"width"`
	Height       *int       `json:"height"`
	FileSize     *int64     `json:"file_size"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type imageDetailResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	MIMEType     string     `json:"mime_type"`
	SourceURL    *string    `json:"source_url"`
	FolderID     *uuid.UUID `json:"folder_id"`
	ThumbnailURL *string    `json:"thumbnail_url"`
	Width        *int       `json:"width"`
	Height       *int       `json:"height"`
	FileSize     *int64     `json:"file_size"`
	ImageURL     string     `json:"image_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type listImagesResponse struct {
	Images     []imageResponse `json:"images"`
	NextCursor *string         `json:"next_cursor"`
}

type completeUploadResponse struct {
	ImageID             uuid.UUID `json:"image_id"`
	SuggestedFolderName *string   `json:"suggested_folder_name"`
	Warning             string    `json:"warning,omitempty"`
}

type acceptSuggestionRequest struct {
	SuggestedFolderName string `json:"suggested_folder_name"`
}

func NewImageHandler(imageUsecase usecase.ImageUsecase, store storage.StorageService, tel *observability.Telemetry) *ImageHandler {
	return &ImageHandler{
		imageUsecase: imageUsecase,
		store:        store,
		tel:          tel,
	}
}

func (h *ImageHandler) InitiateUpload(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.InitiateUpload")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req initiateImageUploadRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := h.imageUsecase.InitiateUpload(ctx, userID, req.Title, req.MIMEType, req.SourceURL, req.FolderID, req.Description)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidImageTitle) || errors.Is(err, usecase.ErrInvalidMIMEType) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to initiate image upload")
	}

	return c.JSON(http.StatusCreated, initiateImageUploadResponse{
		ID:        result.Image.ID,
		UploadURL: result.UploadURL,
		R2Path:    result.Image.R2Path,
	})
}

func (h *ImageHandler) CompleteUpload(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.CompleteUpload")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	result, err := h.imageUsecase.CompleteUpload(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to complete image upload")
	}

	return c.JSON(http.StatusOK, completeUploadResponse{
		ImageID:             result.ImageID,
		SuggestedFolderName: result.SuggestedFolderName,
		Warning:             result.Warning,
	})
}

func (h *ImageHandler) AcceptSuggestion(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.AcceptSuggestion")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	var req acceptSuggestionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.SuggestedFolderName) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "suggested_folder_name is required")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.imageUsecase.AcceptSuggestion(ctx, imageID, userID, req.SuggestedFolderName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to accept folder suggestion")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ImageHandler) ListImages(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListImages")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var folderID *uuid.UUID
	if folderIDParam := c.QueryParam("folder_id"); folderIDParam != "" {
		parsedFolderID, err := uuid.Parse(folderIDParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
		}
		folderID = &parsedFolderID
	}
	unfiled := c.QueryParam("unfiled") == "true"

	limit, cursor, err := parsePaginationParams(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	result, err := h.imageUsecase.ListImages(ctx, userID, usecase.ListImagesParams{
		FolderID: folderID,
		Unfiled:  unfiled,
		Cursor:   cursor,
		Limit:    limit,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list images")
	}

	images := make([]imageResponse, 0, len(result.Images))
	for _, image := range result.Images {
		images = append(images, h.toImageResponse(image))
	}

	var nextCursor *string
	if result.NextCursor != nil {
		encoded := usecase.EncodeCursor(result.NextCursor)
		nextCursor = &encoded
	}

	return c.JSON(http.StatusOK, listImagesResponse{Images: images, NextCursor: nextCursor})
}

func (h *ImageHandler) GetImage(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetImage")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	result, err := h.imageUsecase.GetImage(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get image")
	}

	imageResp := h.toImageResponse(result.Image)
	return c.JSON(http.StatusOK, imageDetailResponse{
		ID:           imageResp.ID,
		Title:        imageResp.Title,
		Description:  imageResp.Description,
		MIMEType:     imageResp.MIMEType,
		SourceURL:    imageResp.SourceURL,
		FolderID:     imageResp.FolderID,
		ThumbnailURL: imageResp.ThumbnailURL,
		Width:        imageResp.Width,
		Height:       imageResp.Height,
		FileSize:     imageResp.FileSize,
		ImageURL:     result.ImageURL,
		CreatedAt:    imageResp.CreatedAt,
		UpdatedAt:    imageResp.UpdatedAt,
	})
}

func (h *ImageHandler) SoftDelete(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.SoftDelete")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.imageUsecase.SoftDelete(ctx, imageID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete image")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ImageHandler) ListTrashed(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListTrashed")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	limit, cursor, err := parsePaginationParams(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	result, err := h.imageUsecase.ListTrashed(ctx, userID, usecase.ListTrashedParams{
		Cursor: cursor,
		Limit:  limit,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list trashed images")
	}

	images := make([]imageResponse, 0, len(result.Images))
	for _, image := range result.Images {
		images = append(images, h.toImageResponse(image))
	}

	var nextCursor *string
	if result.NextCursor != nil {
		encoded := usecase.EncodeCursor(result.NextCursor)
		nextCursor = &encoded
	}

	return c.JSON(http.StatusOK, listImagesResponse{Images: images, NextCursor: nextCursor})
}

func (h *ImageHandler) Restore(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.Restore")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	image, err := h.imageUsecase.Restore(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to restore image")
	}

	return c.JSON(http.StatusOK, h.toImageResponse(image))
}

func (h *ImageHandler) UpdateImage(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.UpdateImage")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req updateImageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title must not be empty")
	}

	params := usecase.UpdateImageParams{
		Title:       req.Title,
		Description: req.Description,
	}

	if len(req.FolderID) > 0 {
		if string(req.FolderID) == "null" {
			params.FolderID = new(*uuid.UUID)
		} else {
			var folderID uuid.UUID
			if err := json.Unmarshal(req.FolderID, &folderID); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid folder_id")
			}
			inner := folderID
			outer := &inner
			params.FolderID = &outer
		}
	}

	image, err := h.imageUsecase.UpdateImage(ctx, imageID, userID, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update image")
	}

	return c.JSON(http.StatusOK, h.toImageResponse(image))
}

func parsePaginationParams(c echo.Context) (limit int, cursor *usecase.ImageCursor, err error) {
	limit = 50
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		parsed, parseErr := strconv.Atoi(limitParam)
		if parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 200 {
		limit = 200
	}

	if cursorParam := c.QueryParam("cursor"); cursorParam != "" {
		cursor, err = usecase.DecodeCursor(cursorParam)
		if err != nil {
			return 0, nil, err
		}
	}
	return limit, cursor, nil
}

func (h *ImageHandler) toImageResponse(image *domain.Image) imageResponse {
	var thumbnailURL *string
	if image.ThumbnailPath != nil {
		url := h.store.CDNUrl(*image.ThumbnailPath)
		thumbnailURL = &url
	}

	return imageResponse{
		ID:           image.ID,
		Title:        image.Title,
		Description:  image.Description,
		MIMEType:     image.MIMEType,
		SourceURL:    image.SourceURL,
		FolderID:     image.FolderID,
		ThumbnailURL: thumbnailURL,
		Width:        image.Width,
		Height:       image.Height,
		FileSize:     image.FileSize,
		CreatedAt:    image.CreatedAt,
		UpdatedAt:    image.UpdatedAt,
	}
}
