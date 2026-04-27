package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	authmw "github.com/devi/bookleaf/internal/middleware"
	"github.com/labstack/echo/v4"
)

type mockUserUsecase struct {
	user     *domain.User
	err      error
	calledID string
}

func (m *mockUserUsecase) GetOrProvision(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserUsecase) GetByID(_ context.Context, kindeID string) (*domain.User, error) {
	m.calledID = kindeID
	return m.user, m.err
}

func TestMeHandler_GetMe_HappyPath(t *testing.T) {
	mockUC := &mockUserUsecase{
		user: &domain.User{
			ID:            "kp_abc123",
			VisionEnabled: false,
		},
	}
	h := NewMeHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	if err := h.GetMe(c); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body["id"] != "kp_abc123" {
		t.Fatalf("expected id kp_abc123, got %v", body["id"])
	}
	if body["vision_enabled"] != false {
		t.Fatalf("expected vision_enabled false, got %v", body["vision_enabled"])
	}
	if mockUC.calledID != "kp_abc123" {
		t.Fatalf("expected GetByID called with kp_abc123, got %s", mockUC.calledID)
	}
}

func TestMeHandler_GetMe_ErrorPath(t *testing.T) {
	mockUC := &mockUserUsecase{
		err: errors.New("db error"),
	}
	h := NewMeHandler(mockUC)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

	err := h.GetMe(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, httpErr.Code)
	}
}
