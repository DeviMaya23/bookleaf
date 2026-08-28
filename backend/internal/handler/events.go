package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/labstack/echo/v4"
)

type Broadcaster interface {
	Subscribe(userID string) chan domain.Event
	Unsubscribe(userID string, ch chan domain.Event)
}

type EventsHandler struct {
	broadcaster Broadcaster
}

func NewEventsHandler(broadcaster Broadcaster) *EventsHandler {
	return &EventsHandler{broadcaster: broadcaster}
}

func (h *EventsHandler) GetEvents(c echo.Context) error {
	userID, ok := middleware.AuthenticatedUserIDFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	ch := h.broadcaster.Subscribe(userID)
	defer h.broadcaster.Unsubscribe(userID, ch)

	ctx := c.Request().Context()
	flusher, _ := c.Response().Writer.(http.Flusher)

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(c.Response(), "data: %s\n\n", data); err != nil {
				return nil
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}
