package handler

import (
	"net/http"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	authmw "github.com/devi/bookleaf/internal/handler/middleware"
)

type mockBroadcaster struct{}

func (m *mockBroadcaster) Subscribe(_ string) chan domain.Event {
	return make(chan domain.Event, 1)
}

func (m *mockBroadcaster) Unsubscribe(_ string, _ chan domain.Event) {}

func TestEventsHandler_GetEvents_UnauthenticatedReturns401(t *testing.T) {
	h := NewEventsHandler(&mockBroadcaster{})
	c, _ := newEchoContext(t, http.MethodGet, "/events", "")
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "")

	err := h.GetEvents(c)

	assertHTTPError(t, err, http.StatusUnauthorized)
}
