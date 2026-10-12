package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type InternalShareUsecase interface {
	GetPublicFoldersByUser(ctx context.Context, userID uuid.UUID) ([]usecase.FolderShareSummary, error)
	GetSharedFolderByFolderID(ctx context.Context, folderID uuid.UUID) (*usecase.SharedFolder, error)
	CheckFolderPublicStatus(ctx context.Context, folderID uuid.UUID) (string, error)
}

type InternalAccountUsecase interface {
	MarkForDeletion(ctx context.Context, idpSubject string) error
}

type InternalUserResolver interface {
	GetByIDPSubject(ctx context.Context, idpSubject string) (*domain.User, error)
}

type InternalFolderImagesUsecase interface {
	ListFolderImages(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]usecase.FolderImageThumbnail, error)
}

type InternalHandler struct {
	shareUsecase        InternalShareUsecase
	accountUsecase      InternalAccountUsecase
	userResolver        InternalUserResolver
	folderImagesUsecase InternalFolderImagesUsecase
	tel                 *observability.Telemetry
}

func NewInternalHandler(shareUsecase InternalShareUsecase, accountUsecase InternalAccountUsecase, userResolver InternalUserResolver, folderImagesUsecase InternalFolderImagesUsecase, tel *observability.Telemetry) *InternalHandler {
	return &InternalHandler{
		shareUsecase:        shareUsecase,
		accountUsecase:      accountUsecase,
		userResolver:        userResolver,
		folderImagesUsecase: folderImagesUsecase,
		tel:                 tel,
	}
}

type folderShareSummaryResponse struct {
	FolderID   string `json:"folder_id"`
	Token      string `json:"token"`
	FolderName string `json:"folder_name"`
}

type listPublicFoldersResponse struct {
	FolderList []folderShareSummaryResponse `json:"folder_list"`
}

func (h *InternalHandler) ListPublicFolders(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListPublicFolders")
	defer span.End()

	idpSubject := c.Param("user_id")

	user, err := h.userResolver.GetByIDPSubject(ctx, idpSubject)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.JSON(http.StatusOK, listPublicFoldersResponse{FolderList: []folderShareSummaryResponse{}})
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to resolve user")
	}

	summaries, err := h.shareUsecase.GetPublicFoldersByUser(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list public folders")
	}

	folderList := make([]folderShareSummaryResponse, len(summaries))
	for i, s := range summaries {
		folderList[i] = folderShareSummaryResponse{
			FolderID:   s.FolderID.String(),
			Token:      s.Token,
			FolderName: s.FolderName,
		}
	}

	return c.JSON(http.StatusOK, listPublicFoldersResponse{FolderList: folderList})
}

func (h *InternalHandler) GetFolderContents(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetFolderContents")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("folder_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	shared, err := h.shareUsecase.GetSharedFolderByFolderID(ctx, folderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get folder contents")
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

func (h *InternalHandler) CheckFolderStatus(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.CheckFolderStatus")
	defer span.End()

	folderID, err := uuid.Parse(c.Param("folder_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	token, err := h.shareUsecase.CheckFolderPublicStatus(ctx, folderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not public")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check folder status")
	}

	return c.JSON(http.StatusOK, shareTokenResponse{Token: token})
}

type folderImageThumbnailResponse struct {
	ImageID      string `json:"image_id"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type listFolderImagesResponse struct {
	Images []folderImageThumbnailResponse `json:"images"`
}

func (h *InternalHandler) ListFolderImages(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListFolderImages")
	defer span.End()

	idpSubject := c.Param("user_id")

	user, err := h.userResolver.GetByIDPSubject(ctx, idpSubject)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to resolve user")
	}

	folderID, err := uuid.Parse(c.Param("folder_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid folder id")
	}

	thumbnails, err := h.folderImagesUsecase.ListFolderImages(ctx, user.ID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "folder not found")
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list folder images")
	}

	images := make([]folderImageThumbnailResponse, len(thumbnails))
	for i, t := range thumbnails {
		images[i] = folderImageThumbnailResponse{
			ImageID:      t.ImageID.String(),
			ThumbnailURL: t.ThumbnailURL,
		}
	}

	return c.JSON(http.StatusOK, listFolderImagesResponse{Images: images})
}

func (h *InternalHandler) DeleteAccount(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DeleteAccount")
	defer span.End()

	idpSubject := c.Param("id")

	if err := h.accountUsecase.MarkForDeletion(ctx, idpSubject); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to schedule account deletion")
	}

	return c.NoContent(http.StatusAccepted)
}
