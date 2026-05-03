package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_GetHealth(t *testing.T) {
	tests := []struct {
		name       string
		dbErr      error
		r2Err      error
		wantStatus string
	}{
		{
			name:       "all healthy",
			wantStatus: "ok",
		},
		{
			name:       "db probe fails",
			dbErr:      errors.New("db timeout"),
			wantStatus: "degraded",
		},
		{
			name:       "r2 probe fails",
			r2Err:      errors.New("r2 unavailable"),
			wantStatus: "degraded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := &HealthHandler{
				dbProbe: func(_ context.Context) error {
					return tt.dbErr
				},
				r2Probe: func(_ context.Context) error {
					return tt.r2Err
				},
			}

			err := h.GetHealth(c)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var resp healthResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tt.wantStatus, resp.Status)

			if tt.dbErr == nil {
				assert.Equal(t, "ok", resp.DB)
			} else {
				assert.Equal(t, tt.dbErr.Error(), resp.DB)
			}

			if tt.r2Err == nil {
				assert.Equal(t, "ok", resp.R2)
			} else {
				assert.Equal(t, tt.r2Err.Error(), resp.R2)
			}
		})
	}
}
