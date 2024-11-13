package metrics

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricsInitialization(t *testing.T) {
	// Ensure that the metrics are registered
	HeapDumpHandled.WithLabelValues("test_tenant").Inc()
	FailedDumps.WithLabelValues("test_tenant").Inc()

	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)

	var foundHeapDumpHandled, foundFailedDumps bool
	for _, mf := range metricFamilies {
		if mf.GetName() == "heap_dump_service_handled_heap_dumps" {
			foundHeapDumpHandled = true
		}
		if mf.GetName() == "heap_dump_service_failed_heap_dumps" {
			foundFailedDumps = true
		}
	}

	assert.True(t, foundHeapDumpHandled, "handled_heap_dumps metric not found")
	assert.True(t, foundFailedDumps, "failed_heap_dumps metric not found")
}

func TestPrometheusServerStartup(t *testing.T) {

	go func() {
		StartMetricServer(21338, "/metrics")
	}()
	// Metrics Server needs a few seconds to start up
	time.Sleep(2 * time.Second)

	request, _ := http.NewRequest(http.MethodGet, "http://localhost:21338/metrics", strings.NewReader(""))
	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Metrics endpoint did not start: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}

}
