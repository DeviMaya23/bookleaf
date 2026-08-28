package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

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

type ShareUsecase interface {
	CreateShare(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) (token string, created bool, err error)
	GetShare(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) (token string, err error)
	DeleteShare(ctx context.Context, folderID uuid.UUID, userID uuid.UUID) error
	GetSharedFolder(ctx context.Context, token string) (*usecase.SharedFolder, error)
	GetSharedFolderInfo(ctx context.Context, token string) (*domain.Folder, error)
}

// FolderExporter is the narrow export capability ShareHandler needs,
// satisfied implicitly by the existing folderUsecase.
type FolderExporter interface {
	ExportFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID, w io.Writer) error
}

type ShareHandler struct {
	shareUsecase   ShareUsecase
	folderExporter FolderExporter
	tel            *observability.Telemetry
}

func NewShareHandler(shareUsecase ShareUsecase, folderExporter FolderExporter, tel *observability.Telemetry) *ShareHandler {
	return &ShareHandler{
		shareUsecase:   shareUsecase,
		folderExporter: folderExporter,
		tel:            tel,
	}
}

type shareTokenResponse struct {
	Token string `json:"token"`
}

type sharedFolderInfo struct {
	Name  string  `json:"name"`
	Notes *string `json:"notes"`
}

type sharedImageResponse struct {
	Title        string  `json:"title"`
	ThumbnailURL *string `json:"thumbnail_url"`
	FullResURL   string  `json:"full_res_url"`
	DownloadURL  string  `json:"download_url"`
	Width        *int    `json:"width"`
	Height       *int    `json:"height"`
}

type sharedFolderResponse struct {
	Folder sharedFolderInfo      `json:"folder"`
	Images []sharedImageResponse `json:"images"`
}

func (h *ShareHandler) CreateShare(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.CreateShare")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	token, created, err := h.shareUsecase.CreateShare(ctx, folderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create folder share")
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	return c.JSON(status, shareTokenResponse{Token: token})
}

func (h *ShareHandler) GetShare(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetShare")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	token, err := h.shareUsecase.GetShare(ctx, folderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder share not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get folder share")
	}

	return c.JSON(http.StatusOK, shareTokenResponse{Token: token})
}

func (h *ShareHandler) DeleteShare(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DeleteShare")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.shareUsecase.DeleteShare(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete folder share")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ShareHandler) GetSharedFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetSharedFolder")
	defer span.End()

	token := c.Param("token")

	shared, err := h.shareUsecase.GetSharedFolder(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "share not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get shared folder")
	}

	images := make([]sharedImageResponse, len(shared.Images))
	for i, img := range shared.Images {
		images[i] = sharedImageResponse{
			Title:        img.Title,
			ThumbnailURL: img.ThumbnailURL,
			FullResURL:   img.FullResURL,
			DownloadURL:  img.DownloadURL,
			Width:        img.Width,
			Height:       img.Height,
		}
	}

	return c.JSON(http.StatusOK, sharedFolderResponse{
		Folder: sharedFolderInfo{Name: shared.Name, Notes: shared.Notes},
		Images: images,
	})
}

func (h *ShareHandler) ExportSharedFolder(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ExportSharedFolder")
	defer span.End()

	token := c.Param("token")

	folder, err := h.shareUsecase.GetSharedFolderInfo(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "share not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get shared folder")
	}

	filename := sanitizeFilename(folder.Name) + ".zip"
	c.Response().Header().Set(echo.HeaderContentType, "application/zip")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", filename))
	c.Response().WriteHeader(http.StatusOK)

	if err := h.folderExporter.ExportFolder(ctx, folder.ID, folder.UserID, c.Response()); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		observability.LoggerFromContext(ctx, h.tel.Logger).Error("export shared folder failed",
			zap.Error(err),
			zap.String("folder_id", folder.ID.String()),
		)
	}

	return nil
}
