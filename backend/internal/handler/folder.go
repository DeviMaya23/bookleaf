package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var invalidFilenameChars = regexp.MustCompile(`[/\\:*?"<>|\x00-\x1f]`)

// sanitizeFilename strips characters invalid in filenames, falling back to
// "export" if nothing usable remains.
func sanitizeFilename(name string) string {
	sanitized := strings.TrimSpace(invalidFilenameChars.ReplaceAllString(name, ""))
	if sanitized == "" {
		return "export"
	}
	return sanitized
}

type FolderUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID, description *string, icon *string) (*domain.Folder, error)
	List(ctx context.Context, userID uuid.UUID) ([]*domain.Folder, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*usecase.FolderDetail, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, params usecase.UpdateFolderParams) (*domain.Folder, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ExportFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID, w io.Writer) error
}

type FolderHandler struct {
	folderUsecase FolderUsecase
	tel           *observability.Telemetry
}

type folderRequest struct {
	Name        string     `json:"name"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Description *string    `json:"description"`
	Icon        *string    `json:"icon"`
}

type updateFolderRequest struct {
	Name        json.RawMessage `json:"name"`
	ParentID    json.RawMessage `json:"parent_id"`
	Description json.RawMessage `json:"description"`
	Icon        json.RawMessage `json:"icon"`
}

type folderResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Icon        *string    `json:"icon"`
	ParentID    *uuid.UUID `json:"parent_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type folderDetailResponse struct {
	folderResponse
	ImageCount int64 `json:"image_count"`
}

func NewFolderHandler(folderUsecase FolderUsecase, tel *observability.Telemetry) *FolderHandler {
	return &FolderHandler{
		folderUsecase: folderUsecase,
		tel:           tel,
	}
}

func (h *FolderHandler) CreateFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.CreateFolder")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req folderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	folder, err := h.folderUsecase.Create(ctx, userID, req.Name, req.ParentID, req.Description, req.Icon)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidFolderName) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder name is required")
		}
		if errors.Is(err, usecase.ErrInvalidFolderIcon) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder icon is not in the allowlist")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create folder")
	}

	return c.JSON(http.StatusCreated, toFolderResponse(folder))
}

func (h *FolderHandler) ListFolders(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListFolders")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	folders, err := h.folderUsecase.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list folders")
	}

	response := make([]folderResponse, 0, len(folders))
	for _, folder := range folders {
		response = append(response, toFolderResponse(folder))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *FolderHandler) GetFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetFolder")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	folderDetail, err := h.folderUsecase.GetByID(ctx, folderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get folder")
	}

	return c.JSON(http.StatusOK, toFolderDetailResponse(folderDetail))
}

func (h *FolderHandler) UpdateFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.UpdateFolder")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req updateFolderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	var params usecase.UpdateFolderParams

	if len(req.Name) > 0 && string(req.Name) != "null" {
		var name string
		if err := json.Unmarshal(req.Name, &name); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid name")
		}
		if strings.TrimSpace(name) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "folder name is required")
		}
		params.Name = &name
	}

	if len(req.ParentID) > 0 {
		if string(req.ParentID) == "null" {
			params.ParentID = new(*uuid.UUID)
		} else {
			var raw string
			if err := json.Unmarshal(req.ParentID, &raw); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid parent_id")
			}
			parsed, err := uuid.Parse(raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid parent_id")
			}
			inner := parsed
			outer := &inner
			params.ParentID = &outer
		}
	}

	if len(req.Description) > 0 {
		if string(req.Description) == "null" {
			params.Description = new(*string)
		} else {
			var description string
			if err := json.Unmarshal(req.Description, &description); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid description")
			}
			inner := description
			outer := &inner
			params.Description = &outer
		}
	}

	if len(req.Icon) > 0 {
		if string(req.Icon) == "null" {
			params.Icon = new(*string)
		} else {
			var icon string
			if err := json.Unmarshal(req.Icon, &icon); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid icon")
			}
			inner := icon
			outer := &inner
			params.Icon = &outer
		}
	}

	folder, err := h.folderUsecase.Update(ctx, folderID, userID, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidFolderName) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder name is required")
		}
		if errors.Is(err, usecase.ErrInvalidFolderIcon) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder icon is not in the allowlist")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update folder")
	}

	return c.JSON(http.StatusOK, toFolderResponse(folder))
}

func (h *FolderHandler) DeleteFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DeleteFolder")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	err = h.folderUsecase.Delete(ctx, folderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete folder")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *FolderHandler) ExportFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ExportFolder")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	folderDetail, err := h.folderUsecase.GetByID(ctx, folderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get folder")
	}

	filename := sanitizeFilename(folderDetail.Folder.Name) + ".zip"
	c.Response().Header().Set(echo.HeaderContentType, "application/zip")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", filename))
	c.Response().WriteHeader(http.StatusOK)

	if err := h.folderUsecase.ExportFolder(ctx, folderID, userID, c.Response()); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		observability.LoggerFromContext(ctx, h.tel.Logger).Error("export folder failed",
			zap.Error(err),
			zap.String("folder_id", folderID.String()),
		)
	}

	return nil
}

func toFolderResponse(folder *domain.Folder) folderResponse {
	return folderResponse{
		ID:          folder.ID,
		Name:        folder.Name,
		Description: folder.Description,
		Icon:        folder.Icon,
		ParentID:    folder.ParentID,
		CreatedAt:   folder.CreatedAt,
		UpdatedAt:   folder.UpdatedAt,
	}
}

func toFolderDetailResponse(folderDetail *usecase.FolderDetail) folderDetailResponse {
	return folderDetailResponse{
		folderResponse: toFolderResponse(folderDetail.Folder),
		ImageCount:     folderDetail.ImageCount,
	}
}
