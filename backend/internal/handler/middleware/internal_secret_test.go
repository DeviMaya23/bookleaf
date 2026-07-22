package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalSecretMiddleware(t *testing.T) {
	const secret = "super-secret"

	tests := []struct {
		name           string
		headerValue    string
		setHeader      bool
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "absent header returns 401 and does not call next",
			setHeader:      false,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "wrong secret value returns 401",
			setHeader:      true,
			headerValue:    "wrong-secret",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "correct secret calls next handler",
			setHeader:      true,
			headerValue:    secret,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/internal/users/kp_abc123/public-folders", nil)
			if tt.setHeader {
				req.Header.Set(internalSecretHeader, tt.headerValue)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			next := func(c echo.Context) error {
				nextCalled = true
				return c.String(http.StatusOK, "ok")
			}

			err := NewInternalSecretMiddleware(secret)(next)(c)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantNextCalled, nextCalled)
		})
	}
}
