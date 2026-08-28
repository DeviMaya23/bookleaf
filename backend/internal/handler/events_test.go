package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/labstack/echo/v4"
)

type mockBroadcaster struct{}

func (m *mockBroadcaster) Subscribe(_ string) chan domain.Event {
	return make(chan domain.Event, 1)
}

func (m *mockBroadcaster) Unsubscribe(_ string, _ chan domain.Event) {}

func TestEventsHandler_GetEvents_UnauthenticatedReturns401(t *testing.T) {
	h := NewEventsHandler(&mockBroadcaster{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No auth key set — simulates unauthenticated request

	err := h.GetEvents(c)

	assertHTTPError(t, err, http.StatusUnauthorized)
}
