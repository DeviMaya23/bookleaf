package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ImageHandler struct {
	imageUsecase usecase.ImageUsecase
	store        storage.StorageService
}

type initiateImageUploadRequest struct {
	Title     string     `json:"title"`
	MIMEType  string     `json:"mime_type"`
	SourceURL *string    `json:"source_url"`
	FolderID  *uuid.UUID `json:"folder_id"`
}

type initiateImageUploadResponse struct {
	ID        uuid.UUID `json:"id"`
	UploadURL string    `json:"upload_url"`
	R2Path    string    `json:"r2_path"`
}

type imageResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	MIMEType     string     `json:"mime_type"`
	SourceURL    *string    `json:"source_url"`
	FolderID     *uuid.UUID `json:"folder_id"`
	ThumbnailURL *string    `json:"thumbnail_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type imageDetailResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	MIMEType     string     `json:"mime_type"`
	SourceURL    *string    `json:"source_url"`
	FolderID     *uuid.UUID `json:"folder_id"`
	ThumbnailURL *string    `json:"thumbnail_url"`
	ImageURL     string     `json:"image_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func NewImageHandler(imageUsecase usecase.ImageUsecase, store storage.StorageService) *ImageHandler {
	return &ImageHandler{
		imageUsecase: imageUsecase,
		store:        store,
	}
}

func (h *ImageHandler) InitiateUpload(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req initiateImageUploadRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	result, err := h.imageUsecase.InitiateUpload(c.Request().Context(), userID, req.Title, req.MIMEType, req.SourceURL, req.FolderID)
	if err != nil {
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
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.imageUsecase.CompleteUpload(c.Request().Context(), imageID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to complete image upload")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ImageHandler) ListImages(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var folderID *uuid.UUID
	folderIDParam := c.QueryParam("folder_id")
	if folderIDParam != "" {
		parsedFolderID, err := uuid.Parse(folderIDParam)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
		}
		folderID = &parsedFolderID
	}

	images, err := h.imageUsecase.ListImages(c.Request().Context(), userID, folderID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list images")
	}

	response := make([]imageResponse, 0, len(images))
	for _, image := range images {
		response = append(response, h.toImageResponse(image))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *ImageHandler) GetImage(c echo.Context) error {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	result, err := h.imageUsecase.GetImage(c.Request().Context(), imageID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get image")
	}

	imageResponse := h.toImageResponse(result.Image)
	return c.JSON(http.StatusOK, imageDetailResponse{
		ID:           imageResponse.ID,
		Title:        imageResponse.Title,
		MIMEType:     imageResponse.MIMEType,
		SourceURL:    imageResponse.SourceURL,
		FolderID:     imageResponse.FolderID,
		ThumbnailURL: imageResponse.ThumbnailURL,
		ImageURL:     result.ImageURL,
		CreatedAt:    imageResponse.CreatedAt,
		UpdatedAt:    imageResponse.UpdatedAt,
	})
}

func (h *ImageHandler) SoftDelete(c echo.Context) error {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.imageUsecase.SoftDelete(c.Request().Context(), imageID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete image")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ImageHandler) ListTrashed(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	images, err := h.imageUsecase.ListTrashed(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list trashed images")
	}

	response := make([]imageResponse, 0, len(images))
	for _, image := range images {
		response = append(response, h.toImageResponse(image))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *ImageHandler) Restore(c echo.Context) error {
	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	image, err := h.imageUsecase.Restore(c.Request().Context(), imageID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to restore image")
	}

	return c.JSON(http.StatusOK, h.toImageResponse(image))
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
		MIMEType:     image.MIMEType,
		SourceURL:    image.SourceURL,
		FolderID:     image.FolderID,
		ThumbnailURL: thumbnailURL,
		CreatedAt:    image.CreatedAt,
		UpdatedAt:    image.UpdatedAt,
	}
}
