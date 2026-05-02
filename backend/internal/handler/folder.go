package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type FolderHandler struct {
	folderUsecase usecase.FolderUsecase
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

func NewFolderHandler(folderUsecase usecase.FolderUsecase) *FolderHandler {
	return &FolderHandler{
		folderUsecase: folderUsecase,
	}
}

func (h *FolderHandler) CreateFolder(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req folderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	folder, err := h.folderUsecase.Create(c.Request().Context(), userID, req.Name, req.ParentID)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidFolderName) {
			return echo.NewHTTPError(http.StatusBadRequest, "folder name is required")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create folder")
	}

	return c.JSON(http.StatusCreated, toFolderResponse(folder))
}

func (h *FolderHandler) ListFolders(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	folders, err := h.folderUsecase.List(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list folders")
	}

	response := make([]folderResponse, 0, len(folders))
	for _, folder := range folders {
		response = append(response, toFolderResponse(folder))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *FolderHandler) GetFolder(c echo.Context) error {
	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	folder, err := h.folderUsecase.GetByID(c.Request().Context(), folderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get folder")
	}

	return c.JSON(http.StatusOK, toFolderResponse(folder))
}

func (h *FolderHandler) UpdateFolder(c echo.Context) error {
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

	folder, err := h.folderUsecase.Update(c.Request().Context(), folderID, userID, req.Name, req.ParentID)
	if err != nil {
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
	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	err = h.folderUsecase.Delete(c.Request().Context(), folderID, userID)
	if err != nil {
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
