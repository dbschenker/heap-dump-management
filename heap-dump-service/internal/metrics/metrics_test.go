package metrics

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPrometheusServerStartup(t *testing.T) {

	go func() {
		StartMetricServer(21338, "/metrics")
	}()

	time.Sleep(1000 * time.Millisecond)

	request, _ := http.NewRequest(http.MethodGet, "http://localhost:21338/metrics", strings.NewReader(""))
	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Errorf("Metrics endpoint did not start: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}

}
