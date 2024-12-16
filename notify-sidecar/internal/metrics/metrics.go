package metrics

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var HeapDumpHandled = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "handled_heap_dumps",
		Namespace: "heap_dump_service",
		Help:      "Number of handled heap dumps",
	},
	[]string{"tenant"},
)

var FailedDumps = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name:      "failed_heap_dumps",
		Namespace: "heap_dump_service",
		Help:      "Number of failed heap dumps",
	},
	[]string{"tenant"},
)

func init() {
	prometheus.MustRegister(HeapDumpHandled)
	prometheus.MustRegister(FailedDumps)
}

func StartMetricServer(port int, path string) {
	http.Handle(path, promhttp.Handler())
	hostAddress := fmt.Sprintf(":%v", port)
	log.WithFields(log.Fields{
		"caller": "StartMetricServer",
	}).Info("Serving Metrics")
	http.ListenAndServe(hostAddress, nil)
}
