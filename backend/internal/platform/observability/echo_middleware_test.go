package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func makeMetricsTestMeter(t *testing.T) (metric.Meter, func() metricdata.ResourceMetrics) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })
	collect := func() metricdata.ResourceMetrics {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &rm))
		return rm
	}
	return mp.Meter("test"), collect
}

func findInt64Sum(rm metricdata.ResourceMetrics, name string) []metricdata.DataPoint[int64] {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				if data, ok := m.Data.(metricdata.Sum[int64]); ok {
					return data.DataPoints
				}
			}
		}
	}
	return nil
}

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

func TestMetricsMiddleware_RequestCount_Success(t *testing.T) {
	meter, collect := makeMetricsTestMeter(t)
	e := echo.New()

	mw := MetricsMiddleware(meter)
	req := httptest.NewRequest(http.MethodGet, "/images", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/images")

	err := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})(c)
	require.NoError(t, err)

	rm := collect()

	countPoints := findInt64Sum(rm, "http.server.request.count")
	require.Len(t, countPoints, 1)
	assert.Equal(t, int64(1), countPoints[0].Value)

	errorPoints := findInt64Sum(rm, "http.server.request.errors")
	assert.Empty(t, errorPoints)
}

func TestMetricsMiddleware_RequestErrors_4xx(t *testing.T) {
	meter, collect := makeMetricsTestMeter(t)
	e := echo.New()

	mw := MetricsMiddleware(meter)
	req := httptest.NewRequest(http.MethodGet, "/images/:id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/images/:id")

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusNotFound)
	})(c)
	require.NoError(t, err)

	rm := collect()

	errorPoints := findInt64Sum(rm, "http.server.request.errors")
	require.Len(t, errorPoints, 1)
	assert.Equal(t, int64(1), errorPoints[0].Value)

	statusClass, ok := errorPoints[0].Attributes.Value(attribute.Key("http.status_class"))
	require.True(t, ok)
	assert.Equal(t, "4xx", statusClass.AsString())
}
