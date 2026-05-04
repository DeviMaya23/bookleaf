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
	"github.com/devi/bookleaf/internal/observability"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserUsecase struct {
	user *domain.User
	err  error
}

func (m *mockUserUsecase) GetOrProvision(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserUsecase) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func TestMeHandler_GetMe(t *testing.T) {
	tests := []struct {
		name          string
		mockUC        *mockUserUsecase
		wantStatus    int
		wantErrStatus int
	}{
		{
			name:       "returns authenticated user",
			mockUC:     &mockUserUsecase{user: &domain.User{ID: "kp_abc123", VisionEnabled: false}},
			wantStatus: http.StatusOK,
		},
		{
			name:          "returns 500 on usecase error",
			mockUC:        &mockUserUsecase{err: errors.New("db error")},
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMeHandler(tt.mockUC, observability.NewTelemetry(nil, nil, nil))
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set(string(authmw.AuthenticatedUserIDContextKey), "kp_abc123")

			err := h.GetMe(c)

			if tt.wantErrStatus != 0 {
				require.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrStatus, httpErr.Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "kp_abc123", body["id"])
			assert.Equal(t, false, body["vision_enabled"])
		})
	}
}
