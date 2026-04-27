package handler

import (
	"net/http"

	"github.com/devi/bookleaf/internal/middleware"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/labstack/echo/v4"
)

type MeHandler struct {
	userUsecase usecase.UserUsecase
}

func NewMeHandler(userUsecase usecase.UserUsecase) *MeHandler {
	return &MeHandler{
		userUsecase: userUsecase,
	}
}

func (h *MeHandler) GetMe(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}

	user, err := h.userUsecase.GetByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch authenticated user")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":             user.ID,
		"vision_enabled": user.VisionEnabled,
	})
}
