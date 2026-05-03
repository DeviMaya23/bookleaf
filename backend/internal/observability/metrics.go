package observability

import (
	"fmt"
	"net/http"
	"os"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewMeterProvider(exporter string) (*sdkmetric.MeterProvider, http.Handler, error) {
	var mp *sdkmetric.MeterProvider
	var handler http.Handler

	switch exporter {
	case "prometheus":
		exp, err := otelprom.New()
		if err != nil {
			return nil, nil, fmt.Errorf("create prometheus exporter: %w", err)
		}
		mp = sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp))
		handler = promhttp.Handler()

	case "gcp":
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		exp, err := mexporter.New(mexporter.WithProjectID(projectID))
		if err != nil {
			return nil, nil, fmt.Errorf("create gcp cloud monitoring exporter: %w", err)
		}
		reader := sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(60*time.Second))
		mp = sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		handler = nil

	default:
		return nil, nil, fmt.Errorf("unknown metrics exporter %q: must be \"prometheus\" or \"gcp\"", exporter)
	}

	otel.SetMeterProvider(mp)
	return mp, handler, nil
}
