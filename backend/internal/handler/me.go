package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
)

type UserUsecase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdatePreferences(ctx context.Context, id uuid.UUID, params usecase.UpdateUserPreferencesParams) (*domain.User, error)
}

type AccountUsecase interface {
	MarkForDeletion(ctx context.Context, idpSubject string) error
}

type CategorisationCountUsecase interface {
	CountThisMonth(ctx context.Context, userID uuid.UUID) (int, error)
}

type MeHandler struct {
	userUsecase           UserUsecase
	accountUsecase        AccountUsecase
	categorisationUsecase CategorisationCountUsecase
	tel                   *observability.Telemetry
}

type updateMeRequest struct {
	VisionEnabled           json.RawMessage `json:"vision_enabled"`
	FolderIconsEnabled      json.RawMessage `json:"folder_icons_enabled"`
	AICategorisationEnabled json.RawMessage `json:"ai_categorisation_enabled"`
}

func NewMeHandler(userUsecase UserUsecase, accountUsecase AccountUsecase, categorisationUsecase CategorisationCountUsecase, tel *observability.Telemetry) *MeHandler {
	return &MeHandler{
		userUsecase:           userUsecase,
		accountUsecase:        accountUsecase,
		categorisationUsecase: categorisationUsecase,
		tel:                   tel,
	}
}

func (h *MeHandler) GetMe(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetMe")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	user, err := h.userUsecase.GetByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch authenticated user")
	}

	count, err := h.categorisationUsecase.CountThisMonth(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch categorisation count")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"vision_enabled":                     user.VisionEnabled,
		"folder_icons_enabled":               user.FolderIconsEnabled,
		"ai_categorisation_enabled":          user.AICategorisationEnabled,
		"ai_categorisation_count_this_month": count,
	})
}

func (h *MeHandler) UpdateMe(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.UpdateMe")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserUUIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	var req updateMeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if len(req.VisionEnabled) == 0 && len(req.FolderIconsEnabled) == 0 && len(req.AICategorisationEnabled) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "vision_enabled, folder_icons_enabled, or ai_categorisation_enabled is required")
	}

	var params usecase.UpdateUserPreferencesParams

	if len(req.VisionEnabled) > 0 {
		var visionEnabled bool
		if err := json.Unmarshal(req.VisionEnabled, &visionEnabled); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "vision_enabled must be a boolean")
		}
		params.VisionEnabled = &visionEnabled
	}

	if len(req.FolderIconsEnabled) > 0 {
		var folderIconsEnabled bool
		if err := json.Unmarshal(req.FolderIconsEnabled, &folderIconsEnabled); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "folder_icons_enabled must be a boolean")
		}
		params.FolderIconsEnabled = &folderIconsEnabled
	}

	if len(req.AICategorisationEnabled) > 0 {
		var aiCategorisationEnabled bool
		if err := json.Unmarshal(req.AICategorisationEnabled, &aiCategorisationEnabled); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "ai_categorisation_enabled must be a boolean")
		}
		params.AICategorisationEnabled = &aiCategorisationEnabled
	}

	user, err := h.userUsecase.UpdatePreferences(ctx, userID, params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
	}

	count, err := h.categorisationUsecase.CountThisMonth(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch categorisation count")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"vision_enabled":                     user.VisionEnabled,
		"folder_icons_enabled":               user.FolderIconsEnabled,
		"ai_categorisation_enabled":          user.AICategorisationEnabled,
		"ai_categorisation_count_this_month": count,
	})
}

func (h *MeHandler) DeleteMe(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.DeleteMe")
	defer span.End()

	idpSubject, ok := middleware.AuthenticatedIDPSubjectFromContext(c)
	if !ok || idpSubject == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated idp subject missing in context")
	}

	if err := h.accountUsecase.MarkForDeletion(ctx, idpSubject); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to schedule account deletion")
	}

	return c.NoContent(http.StatusAccepted)
}
