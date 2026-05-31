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
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type ImageHandler struct {
	imageUsecase usecase.ImageUsecase
	tel          *observability.Telemetry
}

type updateImageRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	FolderID    json.RawMessage `json:"folder_id"`
	SourceURL   json.RawMessage `json:"source_url"`
	Tags        json.RawMessage `json:"tags"`
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
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	MIMEType     string        `json:"mime_type"`
	SourceURL    *string       `json:"source_url"`
	FolderID     *uuid.UUID    `json:"folder_id"`
	Position     *string       `json:"position"`
	ThumbnailURL *string       `json:"thumbnail_url"`
	Width        *int          `json:"width"`
	Height       *int          `json:"height"`
	FileSize     *int64        `json:"file_size"`
	Tags         []tagResponse `json:"tags"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type updateImagePositionRequest struct {
	FolderID *uuid.UUID `json:"folder_id"`
	Position string     `json:"position"`
}

type imageDetailResponse struct {
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	MIMEType     string        `json:"mime_type"`
	SourceURL    *string       `json:"source_url"`
	FolderID     *uuid.UUID    `json:"folder_id"`
	ThumbnailURL *string       `json:"thumbnail_url"`
	Width        *int          `json:"width"`
	Height       *int          `json:"height"`
	FileSize     *int64        `json:"file_size"`
	Tags         []tagResponse `json:"tags"`
	ImageURL     string        `json:"image_url"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type downloadImageResponse struct {
	DownloadURL string `json:"download_url"`
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

func NewImageHandler(imageUsecase usecase.ImageUsecase, tel *observability.Telemetry) *ImageHandler {
	return &ImageHandler{
		imageUsecase: imageUsecase,
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
		ID:        result.ID,
		UploadURL: result.UploadURL,
		R2Path:    result.R2Path,
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
	var tagID *uuid.UUID
	if tagIDParam := c.QueryParam("tag_id"); tagIDParam != "" {
		parsedTagID, err := uuid.Parse(tagIDParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid tag id")
		}
		tagID = &parsedTagID
	}

	var (
		limit  int
		cursor *usecase.ImageCursor
		err    error
	)
	if folderID == nil {
		limit, cursor, err = parsePaginationParams(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
		}
	}

	result, err := h.imageUsecase.ListImages(ctx, userID, usecase.ListImagesParams{
		FolderID: folderID,
		Unfiled:  unfiled,
		TagID:    tagID,
		Cursor:   cursor,
		Limit:    limit,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list images")
	}

	images := make([]imageResponse, 0, len(result.Images))
	for _, item := range result.Images {
		images = append(images, toImageResponse(item))
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

	item := toImageResponse(usecase.ImageItem{Image: result.Image, ThumbnailURL: result.ThumbnailURL})
	return c.JSON(http.StatusOK, imageDetailResponse{
		ID:           item.ID,
		Title:        item.Title,
		Description:  item.Description,
		MIMEType:     item.MIMEType,
		SourceURL:    item.SourceURL,
		FolderID:     item.FolderID,
		ThumbnailURL: item.ThumbnailURL,
		Width:        item.Width,
		Height:       item.Height,
		FileSize:     item.FileSize,
		Tags:         item.Tags,
		ImageURL:     result.ImageURL,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	})
}

func (h *ImageHandler) DownloadImage(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DownloadImage")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	downloadURL, err := h.imageUsecase.DownloadImage(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get image download url")
	}

	return c.JSON(http.StatusOK, downloadImageResponse{DownloadURL: downloadURL})
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
	for _, item := range result.Images {
		images = append(images, toImageResponse(item))
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

	item, err := h.imageUsecase.Restore(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to restore image")
	}

	return c.JSON(http.StatusOK, toImageResponse(*item))
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

	if len(req.SourceURL) > 0 {
		if string(req.SourceURL) == "null" {
			params.SourceURL = new(*string)
		} else {
			var sourceURL string
			if err := json.Unmarshal(req.SourceURL, &sourceURL); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid source_url")
			}
			inner := sourceURL
			outer := &inner
			params.SourceURL = &outer
		}
	}

	if len(req.Tags) > 0 {
		if string(req.Tags) != "null" {
			var tagIDsRaw []string
			if err := json.Unmarshal(req.Tags, &tagIDsRaw); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid tags")
			}
			tagIDs := make([]uuid.UUID, 0, len(tagIDsRaw))
			for _, raw := range tagIDsRaw {
				parsed, err := uuid.Parse(raw)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid tag id")
				}
				tagIDs = append(tagIDs, parsed)
			}
			params.Tags = &tagIDs
		}
	}

	item, err := h.imageUsecase.UpdateImage(ctx, imageID, userID, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update image")
	}

	return c.JSON(http.StatusOK, toImageResponse(*item))
}

func (h *ImageHandler) UpdateImagePosition(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.UpdateImagePosition")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	var req updateImagePositionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.FolderID == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "folder_id is required")
	}
	if req.Position == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "position is required")
	}

	if err := h.imageUsecase.UpdateImagePosition(ctx, imageID, userID, *req.FolderID, req.Position); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found or not in specified folder")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update image position")
	}

	return c.NoContent(http.StatusNoContent)
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

func firstFolderID(imageFolders []domain.ImageFolder) *uuid.UUID {
	if len(imageFolders) == 0 {
		return nil
	}
	id := imageFolders[0].FolderID
	return &id
}

func firstFolderPosition(imageFolders []domain.ImageFolder) *string {
	if len(imageFolders) == 0 {
		return nil
	}
	p := imageFolders[0].Position
	return &p
}

func toImageResponse(item usecase.ImageItem) imageResponse {
	tags := make([]tagResponse, 0, len(item.Image.Tags))
	for _, tag := range item.Image.Tags {
		tags = append(tags, tagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return imageResponse{
		ID:           item.Image.ID,
		Title:        item.Image.Title,
		Description:  item.Image.Description,
		MIMEType:     item.Image.MIMEType,
		SourceURL:    item.Image.SourceURL,
		FolderID:     firstFolderID(item.Image.ImageFolders),
		Position:     firstFolderPosition(item.Image.ImageFolders),
		ThumbnailURL: item.ThumbnailURL,
		Width:        item.Image.Width,
		Height:       item.Image.Height,
		FileSize:     item.Image.FileSize,
		Tags:         tags,
		CreatedAt:    item.Image.CreatedAt,
		UpdatedAt:    item.Image.UpdatedAt,
	}
}
