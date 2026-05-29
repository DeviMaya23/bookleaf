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

type TagHandler struct {
	tagUsecase usecase.TagUsecase
	tel        *observability.Telemetry
}

type createTagRequest struct {
	Name string `json:"name"`
}

type updateTagRequest struct {
	Name string `json:"name"`
}

type tagResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func NewTagHandler(tagUsecase usecase.TagUsecase, tel *observability.Telemetry) *TagHandler {
	return &TagHandler{
		tagUsecase: tagUsecase,
		tel:        tel,
	}
}

func (h *TagHandler) CreateTag(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.CreateTag")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req createTagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	tag, err := h.tagUsecase.Create(ctx, userID, req.Name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidTagName) {
			return echo.NewHTTPError(http.StatusBadRequest, "tag name is required")
		}
		if errors.Is(err, usecase.ErrDuplicateTagName) {
			return echo.NewHTTPError(http.StatusConflict, "tag name already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create tag")
	}

	return c.JSON(http.StatusCreated, toTagResponse(tag))
}

func (h *TagHandler) ListTags(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.ListTags")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	tags, err := h.tagUsecase.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tags")
	}

	response := make([]tagResponse, 0, len(tags))
	for _, tag := range tags {
		response = append(response, toTagResponse(tag))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *TagHandler) UpdateTag(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.UpdateTag")
	defer span.End()

	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tag id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req updateTagRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	tag, err := h.tagUsecase.Update(ctx, tagID, userID, req.Name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, usecase.ErrInvalidTagName) {
			return echo.NewHTTPError(http.StatusBadRequest, "tag name is required")
		}
		if errors.Is(err, usecase.ErrDuplicateTagName) {
			return echo.NewHTTPError(http.StatusConflict, "tag name already exists")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "tag not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update tag")
	}

	return c.JSON(http.StatusOK, toTagResponse(tag))
}

func (h *TagHandler) DeleteTag(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DeleteTag")
	defer span.End()

	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tag id")
	}

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	if err := h.tagUsecase.Delete(ctx, tagID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "tag not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete tag")
	}

	return c.NoContent(http.StatusNoContent)
}

func toTagResponse(tag *domain.Tag) tagResponse {
	createdAt := tag.CreatedAt
	updatedAt := tag.UpdatedAt
	return tagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}
