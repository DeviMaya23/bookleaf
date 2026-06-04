package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

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

type completeUploadResponse struct {
	ImageID             uuid.UUID `json:"image_id"`
	SuggestedFolderName *string   `json:"suggested_folder_name"`
	Warning             string    `json:"warning,omitempty"`
}

type acceptSuggestionRequest struct {
	SuggestedFolderName string `json:"suggested_folder_name"`
}

type UploadUsecase interface {
	InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*usecase.UploadInitResult, error)
	CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*usecase.CompleteUploadResult, error)
	AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error
}

type UploadHandler struct {
	uploadUsecase UploadUsecase
	tel           *observability.Telemetry
}

func NewUploadHandler(uploadUsecase UploadUsecase, tel *observability.Telemetry) *UploadHandler {
	return &UploadHandler{
		uploadUsecase: uploadUsecase,
		tel:           tel,
	}
}

func (h *UploadHandler) InitiateUpload(c echo.Context) error {
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

	result, err := h.uploadUsecase.InitiateUpload(ctx, userID, req.Title, req.MIMEType, req.SourceURL, req.FolderID, req.Description)
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

func (h *UploadHandler) CompleteUpload(c echo.Context) error {
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

	result, err := h.uploadUsecase.CompleteUpload(ctx, imageID, userID)
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

func (h *UploadHandler) AcceptSuggestion(c echo.Context) error {
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

	if err := h.uploadUsecase.AcceptSuggestion(ctx, imageID, userID, req.SuggestedFolderName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to accept folder suggestion")
	}

	return c.NoContent(http.StatusNoContent)
}
