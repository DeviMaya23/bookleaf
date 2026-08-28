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

type TrashUsecase interface {
	SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ListTrashed(ctx context.Context, userID uuid.UUID, params usecase.ListTrashedParams) (*usecase.ListTrashedResult, error)
	Restore(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*usecase.ImageItem, error)
	DeleteFromTrash(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	EmptyTrash(ctx context.Context, userID uuid.UUID) error
	BulkTrash(ctx context.Context, userID uuid.UUID, imageIDs []uuid.UUID) (int, error)
}

type TrashHandler struct {
	trashUsecase TrashUsecase
	tel          *observability.Telemetry
}

func NewTrashHandler(trashUsecase TrashUsecase, tel *observability.Telemetry) *TrashHandler {
	return &TrashHandler{
		trashUsecase: trashUsecase,
		tel:          tel,
	}
}

func (h *TrashHandler) SoftDelete(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.SoftDelete")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.trashUsecase.SoftDelete(ctx, imageID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete image")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TrashHandler) ListTrashed(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListTrashed")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	limit, cursor, err := parsePaginationParams(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	var name *string
	if trimmed := strings.TrimSpace(c.QueryParam("name")); trimmed != "" {
		name = &trimmed
	}

	sortParam := strings.TrimSpace(c.QueryParam("sort"))
	if sortParam != "" && sortParam != "created_at" && sortParam != "title" && sortParam != "deleted_at" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid sort field")
	}
	if sortParam == "" {
		// default sort is by deleted_at desc
		sortParam = "deleted_at"
	}
	sortField := &sortParam

	var direction *string
	if dirParam := strings.TrimSpace(c.QueryParam("direction")); dirParam != "" {
		if dirParam != "asc" && dirParam != "desc" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid direction")
		}
		direction = &dirParam
	} else {
		dispatch := usecase.ResolveSort(sortField, nil)
		defDir := dispatch.DefaultDirection
		direction = &defDir
	}

	result, err := h.trashUsecase.ListTrashed(ctx, userID, usecase.ListTrashedParams{
		Name:      name,
		Sort:      sortField,
		Direction: direction,
		Cursor:    cursor,
		Limit:     limit,
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

func (h *TrashHandler) Restore(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.Restore")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	item, err := h.trashUsecase.Restore(ctx, imageID, userID)
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

func (h *TrashHandler) DeleteFromTrash(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DeleteFromTrash")
	defer span.End()

	imageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.trashUsecase.DeleteFromTrash(ctx, imageID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found in trash")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete image from trash")
	}

	return c.NoContent(http.StatusNoContent)
}

type bulkTrashRequest struct {
	ImageIDs []string `json:"image_ids"`
}

func (h *TrashHandler) BulkTrash(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.BulkTrash")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req bulkTrashRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	imageIDs, err := parseUUIDStrings(req.ImageIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image id")
	}

	count, err := h.trashUsecase.BulkTrash(ctx, userID, imageIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to bulk trash images")
	}

	return c.JSON(http.StatusOK, bulkOperationResponse{SucceededCount: count})
}

func (h *TrashHandler) EmptyTrash(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.EmptyTrash")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.trashUsecase.EmptyTrash(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to empty trash")
	}

	return c.NoContent(http.StatusNoContent)
}
