package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var HeapDumpHandled = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "handled_heap_dumps",
		Namespace: "heap_dump_service",
		Help:      "Number of handled heap dumps",
	},
	[]string{"namespace", "tenant"},
)

func init() {
	prometheus.MustRegister(HeapDumpHandled)
}

func StartMetricServer(port int, path string) {
	http.Handle(path, promhttp.Handler())
	hostAddress := fmt.Sprintf(":%v", port)
	http.ListenAndServe(hostAddress, nil)
}
