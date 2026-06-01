package handler

import (
	"context"
	"net/http"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
)

type UserUsecase interface {
	GetByID(ctx context.Context, kindeID string) (*domain.User, error)
}

type MeHandler struct {
	userUsecase UserUsecase
	tel         *observability.Telemetry
}

func NewMeHandler(userUsecase UserUsecase, tel *observability.Telemetry) *MeHandler {
	return &MeHandler{
		userUsecase: userUsecase,
		tel:         tel,
	}
}

func (h *MeHandler) GetMe(c echo.Context) error {
	ctx, span := h.tel.Tracer.Start(c.Request().Context(), "handler.GetMe")
	defer span.End()

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	user, err := h.userUsecase.GetByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch authenticated user")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":             user.ID,
		"vision_enabled": user.VisionEnabled,
	})
}
