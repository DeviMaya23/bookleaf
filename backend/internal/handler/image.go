package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type ImageUsecase interface {
	ListImages(ctx context.Context, userID string, params usecase.ListImagesParams) (*usecase.ListImagesResult, error)
	ListFolderImages(ctx context.Context, userID string, folderID uuid.UUID, sort *string, direction *string) ([]usecase.ImageItem, error)
	GetImage(ctx context.Context, id uuid.UUID, userID string) (*usecase.ImageDetail, error)
	DownloadImage(ctx context.Context, id uuid.UUID, userID string) (string, error)
	UpdateImage(ctx context.Context, id uuid.UUID, userID string, params usecase.UpdateImageParams) (*usecase.ImageItem, error)
	MoveImageFolder(ctx context.Context, imageID uuid.UUID, userID string, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) (*usecase.ImageItem, error)
	UpdateImagePosition(ctx context.Context, imageID uuid.UUID, userID string, folderID uuid.UUID, position string) error
}

type ImageHandler struct {
	imageUsecase ImageUsecase
	tel          *observability.Telemetry
}

type updateImageRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	FolderIDs   json.RawMessage `json:"folder_ids"`
	SourceURL   json.RawMessage `json:"source_url"`
	Tags        json.RawMessage `json:"tags"`
}

type moveImageFolderRequest struct {
	FromFolderID json.RawMessage `json:"from_folder_id"`
	ToFolderID   json.RawMessage `json:"to_folder_id"`
}

type imageResponse struct {
	ID           uuid.UUID     `json:"id"`
	Title        string        `json:"title"`
	Description  *string       `json:"description"`
	MIMEType     string        `json:"mime_type"`
	SourceURL    *string       `json:"source_url"`
	FolderIDs    []uuid.UUID   `json:"folder_ids"`
	ThumbnailURL *string       `json:"thumbnail_url"`
	Width        *int          `json:"width"`
	Height       *int          `json:"height"`
	FileSize     *int64        `json:"file_size"`
	Tags         []tagResponse `json:"tags"`
	Position     *string       `json:"position"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type updateImagePositionRequest struct {
	FolderID *uuid.UUID `json:"folder_id"`
	Position string     `json:"position"`
}

type imageDetailResponse struct {
	ID                  uuid.UUID     `json:"id"`
	Title               string        `json:"title"`
	Description         *string       `json:"description"`
	MIMEType            string        `json:"mime_type"`
	SourceURL           *string       `json:"source_url"`
	FolderIDs           []uuid.UUID   `json:"folder_ids"`
	ThumbnailURL        *string       `json:"thumbnail_url"`
	Width               *int          `json:"width"`
	Height              *int          `json:"height"`
	FileSize            *int64        `json:"file_size"`
	Tags                []tagResponse `json:"tags"`
	Position            *string       `json:"position"`
	ImageURL            string        `json:"image_url"`
	SuggestedFolderName *string       `json:"suggested_folder_name"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

type downloadImageResponse struct {
	DownloadURL string `json:"download_url"`
}

type listImagesResponse struct {
	Images     []imageResponse `json:"images"`
	NextCursor *string         `json:"next_cursor"`
}

func NewImageHandler(imageUsecase ImageUsecase, tel *observability.Telemetry) *ImageHandler {
	return &ImageHandler{
		imageUsecase: imageUsecase,
		tel:          tel,
	}
}

func (h *ImageHandler) ListImages(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListImages")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	unfiled := c.QueryParam("unfiled") == "true"

	var name *string
	if trimmed := strings.TrimSpace(c.QueryParam("name")); trimmed != "" {
		name = &trimmed
	}

	folderIDs, err := parseUUIDListParam(c.QueryParam("folder_ids"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder ids")
	}

	tagIDs, err := parseUUIDListParam(c.QueryParam("tag_ids"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tag ids")
	}

	mimeTypes := parseStringListParam(c.QueryParam("mime_types"))

	var sortField *string
	if sortParam := strings.TrimSpace(c.QueryParam("sort")); sortParam != "" {
		if sortParam != "created_at" && sortParam != "title" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid sort field")
		}
		sortField = &sortParam
	}

	var direction *string
	if dirParam := strings.TrimSpace(c.QueryParam("direction")); dirParam != "" {
		if sortField != nil {
			if dirParam != "asc" && dirParam != "desc" {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid direction")
			}
			direction = &dirParam
		}
	} else if sortField != nil {
		dispatch := usecase.ResolveSort(sortField, nil)
		defDir := dispatch.DefaultDirection
		direction = &defDir
	}

	limit, cursor, err := parsePaginationParams(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	result, err := h.imageUsecase.ListImages(ctx, userID, usecase.ListImagesParams{
		Unfiled:   unfiled,
		FolderIDs: folderIDs,
		TagIDs:    tagIDs,
		MIMETypes: mimeTypes,
		Name:      name,
		Sort:      sortField,
		Direction: direction,
		Cursor:    cursor,
		Limit:     limit,
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

func (h *ImageHandler) ListFolderImages(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListFolderImages")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var sortField *string
	if sortParam := strings.TrimSpace(c.QueryParam("sort")); sortParam != "" {
		if sortParam != "created_at" && sortParam != "title" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid sort field")
		}
		sortField = &sortParam
	}

	var direction *string
	if dirParam := strings.TrimSpace(c.QueryParam("direction")); dirParam != "" {
		if sortField != nil {
			if dirParam != "asc" && dirParam != "desc" {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid direction")
			}
			direction = &dirParam
		}
	} else if sortField != nil {
		dispatch := usecase.ResolveSort(sortField, nil)
		defDir := dispatch.DefaultDirection
		direction = &defDir
	}

	items, err := h.imageUsecase.ListFolderImages(ctx, userID, folderID, sortField, direction)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list folder images")
	}

	images := make([]imageResponse, 0, len(items))
	for _, item := range items {
		images = append(images, toImageResponse(item))
	}

	return c.JSON(http.StatusOK, images)
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
		ID:                  item.ID,
		Title:               item.Title,
		Description:         item.Description,
		MIMEType:            item.MIMEType,
		SourceURL:           item.SourceURL,
		FolderIDs:           item.FolderIDs,
		ThumbnailURL:        item.ThumbnailURL,
		Width:               item.Width,
		Height:               item.Height,
		FileSize:            item.FileSize,
		Tags:                item.Tags,
		ImageURL:            result.ImageURL,
		SuggestedFolderName: result.SuggestedFolderName,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
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

	if len(req.FolderIDs) > 0 && string(req.FolderIDs) != "null" {
		var folderIDsRaw []string
		if err := json.Unmarshal(req.FolderIDs, &folderIDsRaw); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid folder_ids")
		}
		folderIDs := make([]uuid.UUID, 0, len(folderIDsRaw))
		for _, raw := range folderIDsRaw {
			parsed, err := uuid.Parse(raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
			}
			folderIDs = append(folderIDs, parsed)
		}
		params.FolderIDs = &folderIDs
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

func (h *ImageHandler) MoveImageFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.MoveImageFolder")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req moveImageFolderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if len(req.FromFolderID) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "from_folder_id is required")
	}
	if len(req.ToFolderID) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "to_folder_id is required")
	}

	parseFolderID := func(raw json.RawMessage, field string) (*uuid.UUID, error) {
		if string(raw) == "null" {
			return nil, nil
		}
		var folderID uuid.UUID
		if err := json.Unmarshal(raw, &folderID); err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid "+field)
		}
		return &folderID, nil
	}

	fromFolderID, err := parseFolderID(req.FromFolderID, "from_folder_id")
	if err != nil {
		return err
	}
	toFolderID, err := parseFolderID(req.ToFolderID, "to_folder_id")
	if err != nil {
		return err
	}

	item, err := h.imageUsecase.MoveImageFolder(ctx, imageID, userID, fromFolderID, toFolderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to move image folder")
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

// parseUUIDListParam splits a comma-separated query param into UUIDs, ignoring empty
// segments (e.g. from a trailing comma). Returns an error if any non-empty segment fails to parse.
func parseUUIDListParam(raw string) ([]uuid.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var ids []uuid.UUID
	for _, segment := range strings.Split(raw, ",") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		id, err := uuid.Parse(segment)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// parseStringListParam splits a comma-separated query param into non-empty strings,
// ignoring empty segments (e.g. from a trailing comma).
func parseStringListParam(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var values []string
	for _, segment := range strings.Split(raw, ",") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		values = append(values, segment)
	}
	return values
}

func toImageResponse(item usecase.ImageItem) imageResponse {
	tags := make([]tagResponse, 0, len(item.Image.Tags))
	for _, tag := range item.Image.Tags {
		tags = append(tags, tagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	folderIDs := make([]uuid.UUID, 0, len(item.Image.ImageFolders))
	for _, f := range item.Image.ImageFolders {
		folderIDs = append(folderIDs, f.FolderID)
	}
	return imageResponse{
		ID:           item.Image.ID,
		Title:        item.Image.Title,
		Description:  item.Image.Description,
		MIMEType:     item.Image.MIMEType,
		SourceURL:    item.Image.SourceURL,
		FolderIDs:    folderIDs,
		ThumbnailURL: item.ThumbnailURL,
		Width:        item.Image.Width,
		Height:       item.Image.Height,
		FileSize:     item.Image.FileSize,
		Tags:         tags,
		Position:     item.FolderPosition,
		CreatedAt:    item.Image.CreatedAt,
		UpdatedAt:    item.Image.UpdatedAt,
	}
}
