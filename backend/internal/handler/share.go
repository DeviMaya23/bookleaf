package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type ShareUsecase interface {
	CreateShare(ctx context.Context, folderID uuid.UUID, userID string) (token string, created bool, err error)
	GetShare(ctx context.Context, folderID uuid.UUID, userID string) (token string, err error)
	DeleteShare(ctx context.Context, folderID uuid.UUID, userID string) error
	GetSharedFolder(ctx context.Context, token string) (*usecase.SharedFolder, error)
}

type ShareHandler struct {
	shareUsecase ShareUsecase
	tel          *observability.Telemetry
}

func NewShareHandler(shareUsecase ShareUsecase, tel *observability.Telemetry) *ShareHandler {
	return &ShareHandler{
		shareUsecase: shareUsecase,
		tel:          tel,
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

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
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

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
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

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
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
		}
	}

	return c.JSON(http.StatusOK, sharedFolderResponse{
		Folder: sharedFolderInfo{Name: shared.Name, Notes: shared.Notes},
		Images: images,
	})
}
