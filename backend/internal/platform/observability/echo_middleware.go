package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TraceMiddleware(tracer trace.Tracer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			propagator := propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)
			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			route := c.Path()
			if route == "" {
				route = req.URL.Path
			}
			spanName := fmt.Sprintf("%s %s", req.Method, route)

			ctx, span := tracer.Start(ctx, spanName)
			defer span.End()

			c.SetRequest(req.WithContext(ctx))

			err := next(c)

			statusCode := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				} else {
					statusCode = http.StatusInternalServerError
				}
			}

			span.SetAttributes(attribute.Int("http.status_code", statusCode))

			return err
		}
	}
}

func LoggingMiddleware(tel *Telemetry, userIDFromCtx func(echo.Context) (string, bool)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			route := c.Path()
			if route == "" {
				route = c.Request().URL.Path
			}

			statusCode := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				} else {
					statusCode = http.StatusInternalServerError
				}
			}

			userID, _ := userIDFromCtx(c)
			logger := LoggerFromContext(c.Request().Context(), tel.Logger)
			logger.Info("request",
				zap.String("user_id", userID),
				zap.String("http.request.method", c.Request().Method),
				zap.String("http.route", route),
				zap.Int("http.response.status_code", statusCode),
				zap.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
			)

			return err
		}
	}
}

func MetricsMiddleware(meter metric.Meter) echo.MiddlewareFunc {
	duration, _ := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of HTTP server requests in milliseconds"),
	)
	activeRequests, _ := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of in-flight HTTP server requests"),
	)
	requestCount, _ := meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP server requests"),
	)
	requestErrors, _ := meter.Int64Counter(
		"http.server.request.errors",
		metric.WithDescription("Total number of HTTP server requests that resulted in a 4xx or 5xx response"),
	)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			start := time.Now()

			route := c.Path()
			if route == "" {
				route = req.URL.Path
			}

			baseAttrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(req.Method),
				attribute.String("http.route", route),
			}

			activeRequests.Add(req.Context(), 1, metric.WithAttributes(baseAttrs...))
			defer func() {
				statusCode := c.Response().Status
				attrs := append(baseAttrs, semconv.HTTPResponseStatusCodeKey.Int(statusCode))
				activeRequests.Add(req.Context(), -1, metric.WithAttributes(baseAttrs...))
				duration.Record(req.Context(), float64(time.Since(start).Milliseconds()), metric.WithAttributes(attrs...))
				requestCount.Add(req.Context(), 1, metric.WithAttributes(attrs...))
				if statusCode >= 400 {
					statusClass := "4xx"
					if statusCode >= 500 {
						statusClass = "5xx"
					}
					errorAttrs := append(baseAttrs, attribute.String("http.status_class", statusClass))
					requestErrors.Add(req.Context(), 1, metric.WithAttributes(errorAttrs...))
				}
			}()

			return next(c)
		}
	}
}
