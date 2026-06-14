package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaintenanceMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		cfg            config.MaintenanceConfig
		bypassHeader   string
		authHeader     string
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "maintenance disabled passes through",
			cfg:            config.MaintenanceConfig{Enabled: false},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "maintenance enabled, no bypass header returns 503",
			cfg:            config.MaintenanceConfig{Enabled: true},
			wantStatus:     http.StatusServiceUnavailable,
			wantNextCalled: false,
		},
		{
			name:           "maintenance enabled, no Authorization header returns 503 not 401",
			cfg:            config.MaintenanceConfig{Enabled: true},
			authHeader:     "",
			wantStatus:     http.StatusServiceUnavailable,
			wantNextCalled: false,
		},
		{
			name:           "maintenance enabled with matching bypass header passes through",
			cfg:            config.MaintenanceConfig{Enabled: true, BypassToken: "secret"},
			bypassHeader:   "secret",
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "maintenance enabled with non-matching bypass header returns 503",
			cfg:            config.MaintenanceConfig{Enabled: true, BypassToken: "secret"},
			bypassHeader:   "wrong",
			wantStatus:     http.StatusServiceUnavailable,
			wantNextCalled: false,
		},
		{
			name:           "maintenance enabled with unset bypass token returns 503 regardless of bypass header",
			cfg:            config.MaintenanceConfig{Enabled: true, BypassToken: ""},
			bypassHeader:   "anything",
			wantStatus:     http.StatusServiceUnavailable,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.bypassHeader != "" {
				req.Header.Set("X-Bookleaf-Bypass", tt.bypassHeader)
			}
			if tt.authHeader != "" {
				req.Header.Set(echo.HeaderAuthorization, tt.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			next := func(c echo.Context) error {
				nextCalled = true
				return c.String(http.StatusOK, "ok")
			}

			err := NewMaintenanceMiddleware(tt.cfg)(next)(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantNextCalled, nextCalled)

			if tt.wantStatus == http.StatusServiceUnavailable {
				assert.Equal(t, "true", rec.Header().Get("X-Bookleaf-Maintenance"))
				assert.JSONEq(t, `{"error":"maintenance"}`, rec.Body.String())
			}
		})
	}
}
