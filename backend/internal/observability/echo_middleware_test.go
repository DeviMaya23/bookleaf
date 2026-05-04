package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func makeLoggingTestTelemetry(t *testing.T) (*Telemetry, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zapcore.InfoLevel)
	tel := NewTelemetry(nil, nil, nil)
	tel.Logger = zap.New(core)
	return tel, logs
}

func alwaysUserID(id string) func(echo.Context) (string, bool) {
	return func(echo.Context) (string, bool) { return id, true }
}

func TestLoggingMiddleware_Success(t *testing.T) {
	tel, logs := makeLoggingTestTelemetry(t)
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := LoggingMiddleware(tel, alwaysUserID("kp_abc123"))
	req := httptest.NewRequest(http.MethodGet, "/images", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/images")

	err := mw(handler)(c)
	require.NoError(t, err)

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "request", entry.Message)

	fields := entry.ContextMap()
	assert.Equal(t, "kp_abc123", fields["user_id"])
	assert.Equal(t, http.MethodGet, fields["http.request.method"])
	assert.Equal(t, "/images", fields["http.route"])
	assert.EqualValues(t, http.StatusOK, fields["http.response.status_code"])
	_, hasDuration := fields["duration_ms"]
	assert.True(t, hasDuration)
}

func TestLoggingMiddleware_ErrorResponse(t *testing.T) {
	tel, logs := makeLoggingTestTelemetry(t)
	e := echo.New()

	handler := func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "boom")
	}

	mw := LoggingMiddleware(tel, alwaysUserID("kp_xyz"))
	req := httptest.NewRequest(http.MethodPost, "/images", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/images")

	err := mw(handler)(c)
	require.Error(t, err)

	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	assert.EqualValues(t, http.StatusInternalServerError, fields["http.response.status_code"])
	assert.Equal(t, http.MethodPost, fields["http.request.method"])
}
