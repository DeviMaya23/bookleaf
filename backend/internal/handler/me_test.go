package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserUsecase struct {
	user *domain.User
	err  error
}

func (m *mockUserUsecase) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

// --- GetMe ---

func TestMeHandler_GetMe_ReturnsUser(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/me", "")

	err := h.GetMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "kp_abc123", resp["id"])
	assert.Equal(t, true, resp["vision_enabled"])
}

func TestMeHandler_GetMe_GenericError(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/me", "")

	err := h.GetMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}
