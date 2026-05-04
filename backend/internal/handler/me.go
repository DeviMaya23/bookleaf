package handler

import (
	"net/http"

	"github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
)

type MeHandler struct {
	userUsecase usecase.UserUsecase
	tel         *observability.Telemetry
}

func NewMeHandler(userUsecase usecase.UserUsecase, tel *observability.Telemetry) *MeHandler {
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
