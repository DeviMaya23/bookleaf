package handler

import (
	"net/http"

	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SuggestionHandler struct {
	suggestionUsecase usecase.SuggestionUsecase
}

func NewSuggestionHandler(suggestionUsecase usecase.SuggestionUsecase) *SuggestionHandler {
	return &SuggestionHandler{
		suggestionUsecase: suggestionUsecase,
	}
}

func (h *SuggestionHandler) GetSuggestion(c echo.Context) error {

	// dummy

	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "authenticated user id missing in context")
	}
	str, err := h.suggestionUsecase.CategoriseImage(c.Request().Context(), userID, uuid.MustParse("0a34a56d-f443-489a-867b-dc884f83178f"))
	if err != nil {
		return c.JSON(500, err.Error())
	}
	return c.JSON(200, str)
}
