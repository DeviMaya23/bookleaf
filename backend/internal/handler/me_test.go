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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserUsecase struct {
	user *domain.User
	err  error

	lastUpdateID     uuid.UUID
	lastUpdateParams usecase.UpdateUserPreferencesParams
}

func (m *mockUserUsecase) GetByID(_ context.Context, _ uuid.UUID) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockUserUsecase) UpdatePreferences(_ context.Context, id uuid.UUID, params usecase.UpdateUserPreferencesParams) (*domain.User, error) {
	m.lastUpdateID = id
	m.lastUpdateParams = params
	return m.user, m.err
}

type mockAccountUsecase struct {
	err                    error
	markForDeletionCalls   int
	lastMarkForDeletionUser string
}

func (m *mockAccountUsecase) MarkForDeletion(_ context.Context, userID string) error {
	m.markForDeletionCalls++
	m.lastMarkForDeletionUser = userID
	return m.err
}

type mockCategorisationCountUsecase struct {
	count int
	err   error
}

func (m *mockCategorisationCountUsecase) CountThisMonth(_ context.Context, _ uuid.UUID) (int, error) {
	return m.count, m.err
}

func newMeHandler(userUc *mockUserUsecase, countUc *mockCategorisationCountUsecase) *MeHandler {
	if countUc == nil {
		countUc = &mockCategorisationCountUsecase{}
	}
	return NewMeHandler(userUc, &mockAccountUsecase{}, countUc, observability.NewTelemetry(nil, nil, nil))
}

// --- GetMe ---

func TestMeHandler_GetMe_ReturnsUser(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{user: &domain.User{ID: uuid.New(), VisionEnabled: true, FolderIconsEnabled: true, AICategorisationEnabled: true}}, &mockCategorisationCountUsecase{count: 14})
	c, rec := newEchoContext(t, http.MethodGet, "/me", "")

	err := h.GetMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["vision_enabled"])
	assert.Equal(t, true, resp["folder_icons_enabled"])
	assert.Equal(t, true, resp["ai_categorisation_enabled"])
	assert.Equal(t, float64(14), resp["ai_categorisation_count_this_month"])
}

func TestMeHandler_GetMe_GenericError(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{err: errors.New("db error")}, nil)
	c, _ := newEchoContext(t, http.MethodGet, "/me", "")

	err := h.GetMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

// --- UpdateMe ---

func TestMeHandler_UpdateMe_EnablesVision(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: uuid.New(), VisionEnabled: true}}
	h := newMeHandler(uc, nil)
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": true}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.VisionEnabled)
	assert.True(t, *uc.lastUpdateParams.VisionEnabled)
	assert.NotEqual(t, uuid.Nil, uc.lastUpdateID)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["vision_enabled"])
}

func TestMeHandler_UpdateMe_DisablesVision(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: uuid.New(), VisionEnabled: false}}
	h := newMeHandler(uc, nil)
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": false}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.VisionEnabled)
	assert.False(t, *uc.lastUpdateParams.VisionEnabled)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["vision_enabled"])
}

func TestMeHandler_UpdateMe_DisablesFolderIcons(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: uuid.New(), VisionEnabled: true, FolderIconsEnabled: false}}
	h := newMeHandler(uc, nil)
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"folder_icons_enabled": false}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.FolderIconsEnabled)
	assert.False(t, *uc.lastUpdateParams.FolderIconsEnabled)
	assert.Nil(t, uc.lastUpdateParams.VisionEnabled)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["folder_icons_enabled"])
}

func TestMeHandler_UpdateMe_NonBooleanFolderIconsEnabled(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{}, nil)
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"folder_icons_enabled": "yes"}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestMeHandler_UpdateMe_MissingField(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{}, nil)
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestMeHandler_UpdateMe_NonBooleanField(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{}, nil)
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": "yes"}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestMeHandler_UpdateMe_UsecaseError(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{err: errors.New("db error")}, nil)
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"vision_enabled": true}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestMeHandler_UpdateMe_EnablesAICategorisation(t *testing.T) {
	uc := &mockUserUsecase{user: &domain.User{ID: uuid.New(), AICategorisationEnabled: true}}
	h := newMeHandler(uc, &mockCategorisationCountUsecase{count: 5})
	c, rec := newEchoContext(t, http.MethodPatch, "/me", `{"ai_categorisation_enabled": true}`)

	err := h.UpdateMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, uc.lastUpdateParams.AICategorisationEnabled)
	assert.True(t, *uc.lastUpdateParams.AICategorisationEnabled)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["ai_categorisation_enabled"])
	assert.Equal(t, float64(5), resp["ai_categorisation_count_this_month"])
}

func TestMeHandler_UpdateMe_NonBooleanAICategorisationEnabled(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{}, nil)
	c, _ := newEchoContext(t, http.MethodPatch, "/me", `{"ai_categorisation_enabled": "yes"}`)

	err := h.UpdateMe(c)

	assertHTTPError(t, err, http.StatusBadRequest)
}

// --- DeleteMe ---

func TestMeHandler_DeleteMe_Success(t *testing.T) {
	accountUsecase := &mockAccountUsecase{}
	h := NewMeHandler(&mockUserUsecase{}, accountUsecase, &mockCategorisationCountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, rec := newEchoContext(t, http.MethodDelete, "/me", "")

	err := h.DeleteMe(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Equal(t, 1, accountUsecase.markForDeletionCalls)
	assert.Equal(t, "kp_abc123", accountUsecase.lastMarkForDeletionUser)
}

func TestMeHandler_DeleteMe_MissingAuthContext(t *testing.T) {
	h := newMeHandler(&mockUserUsecase{}, nil)
	c, _ := newEchoContext(t, http.MethodDelete, "/me", "")
	c.Set(string(authmw.AuthenticatedIDPSubjectContextKey), "")

	err := h.DeleteMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}

func TestMeHandler_DeleteMe_GenericError(t *testing.T) {
	h := NewMeHandler(&mockUserUsecase{}, &mockAccountUsecase{err: errors.New("db error")}, &mockCategorisationCountUsecase{}, observability.NewTelemetry(nil, nil, nil))
	c, _ := newEchoContext(t, http.MethodDelete, "/me", "")

	err := h.DeleteMe(c)

	assertHTTPError(t, err, http.StatusInternalServerError)
}
