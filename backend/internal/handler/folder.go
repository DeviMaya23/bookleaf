package handler

import (
	"errors"
	"net/http"
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

type FolderHandler struct {
	folderUsecase usecase.FolderUsecase
	tel           *observability.Telemetry
}

type folderRequest struct {
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type folderResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func NewFolderHandler(folderUsecase usecase.FolderUsecase, tel *observability.Telemetry) *FolderHandler {
	return &FolderHandler{
		folderUsecase: folderUsecase,
		tel:           tel,
	}
}

func (h *FolderHandler) CreateFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.CreateFolder")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req folderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	folder, err := h.folderUsecase.Create(ctx, userID, req.Name, req.ParentID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidFolderName) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder name is required")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create folder")
	}

	return c.JSON(http.StatusCreated, toFolderResponse(folder))
}

func (h *FolderHandler) ListFolders(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListFolders")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
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

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	folder, err := h.folderUsecase.GetByID(ctx, folderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get folder")
	}

	return c.JSON(http.StatusOK, toFolderResponse(folder))
}

func (h *FolderHandler) UpdateFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.UpdateFolder")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req folderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	folder, err := h.folderUsecase.Update(ctx, folderID, userID, req.Name, req.ParentID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidFolderName) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder name is required")
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

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
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

func toFolderResponse(folder *domain.Folder) folderResponse {
	return folderResponse{
		ID:        folder.ID,
		Name:      folder.Name,
		ParentID:  folder.ParentID,
		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}
}
