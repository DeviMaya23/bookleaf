package handler

import (
	"context"

	"github.com/labstack/echo/v4"
)

type SuggestionUsecase interface {
	Test(ctx context.Context, userID string, input string) (string, error)
	Test2(ctx context.Context) string
}

type SuggestionHandler struct {
	suggestionUsecase SuggestionUsecase
}

func NewSuggestionHandler(suggestionUsecase SuggestionUsecase) *SuggestionHandler {
	return &SuggestionHandler{
		suggestionUsecase: suggestionUsecase,
	}
}

func (h *SuggestionHandler) GetSuggestion(c echo.Context) error {

	// dummy
	str, err := h.suggestionUsecase.Test(c.Request().Context(), "user123", "input")
	if err != nil {
		return c.JSON(500, err.Error())
	}
	return c.JSON(200, str)
}
