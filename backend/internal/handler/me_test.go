package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/devi/bookleaf/internal/domain"
	authmw "github.com/devi/bookleaf/internal/handler/middleware"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserUsecase struct {
	user *domain.User
	err  error

	lastUpdateID     string
	lastUpdateParams usecase.UpdateUserPreferencesParams
}

func (m *mockUserUsecase) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserUsecase) UpdatePreferences(_ context.Context, id string, params usecase.UpdateUserPreferencesParams) (*domain.User, error) {
	m.lastUpdateID = id
	m.lastUpdateParams = params
	return m.user, m.err
}

type mockAccountUsecase struct {
	err            error
	deleteCalls    int
	lastDeleteUser string
}

func (m *mockAccountUsecase) DeleteAccount(_ context.Context, userID string) error {
	m.deleteCalls++
	m.lastDeleteUser = userID
	return m.err
}

// --- GetMe ---

func TestMeHandler_GetMe_ReturnsUser(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{user: &domain.User{ID: "kp_abc123", VisionEnabled: true, FolderIconsEnabled: true}}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodGet, "/me", "")

	err := h.GetMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "kp_abc123", resp["id"])
	assert.Equal(t, true, resp["vision_enabled"])
	assert.Equal(t, true, resp["folder_icons_enabled"])
}

func TestMeHandler_GetMe_GenericError(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{err: errors.New("db error")}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodGet, "/me", "")

	err := h.GetMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- UpdateMe ---

func TestMeHandler_UpdateMe_EnablesVision(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: "kp_abc123", VisionEnabled: true}}
	h := NewMeHandler(uc, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": true}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.VisionEnabled)
	assert.True(t, *uc.lastUpdateParams.VisionEnabled)
	assert.Equal(t, "kp_abc123", uc.lastUpdateID)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "kp_abc123", resp["id"])
	assert.Equal(t, true, resp["vision_enabled"])
}

func TestMeHandler_UpdateMe_DisablesVision(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: "kp_abc123", VisionEnabled: false}}
	h := NewMeHandler(uc, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": false}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.VisionEnabled)
	assert.False(t, *uc.lastUpdateParams.VisionEnabled)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "kp_abc123", resp["id"])
	assert.Equal(t, false, resp["vision_enabled"])
}

func TestMeHandler_UpdateMe_DisablesFolderIcons(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: "kp_abc123", VisionEnabled: true, FolderIconsEnabled: false}}
	h := NewMeHandler(uc, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"folder_icons_enabled": false}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.FolderIconsEnabled)
	assert.False(t, *uc.lastUpdateParams.FolderIconsEnabled)
	assert.Nil(t, uc.lastUpdateParams.VisionEnabled)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "kp_abc123", resp["id"])
	assert.Equal(t, false, resp["folder_icons_enabled"])
}

func TestMeHandler_UpdateMe_NonBooleanFolderIconsEnabled(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"folder_icons_enabled": "yes"}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestMeHandler_UpdateMe_MissingField(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestMeHandler_UpdateMe_NonBooleanField(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": "yes"}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestMeHandler_UpdateMe_UsecaseError(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{err: errors.New("db error")}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": true}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- DeleteMe ---

func TestMeHandler_DeleteMe_Success(t *testing.T) {
	accountUsecase := &mockAccountUsecase{}
	h := NewMeHandler(&mockUserUsecase{}, accountUsecase, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/me", "")

	err := h.DeleteMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, 1, accountUsecase.deleteCalls)
	assert.Equal(t, "kp_abc123", accountUsecase.lastDeleteUser)
}

func TestMeHandler_DeleteMe_MissingAuthContext(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{}, &mockAccountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/me", "")
	c.Set(string(authmw.AuthenticatedUserIDContextKey), "")

	err := h.DeleteMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestMeHandler_DeleteMe_GenericError(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{}, &mockAccountUsecase{err: errors.New("db error")}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/me", "")

	err := h.DeleteMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}
