package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Metrics struct {
	E2EDelay     *prometheus.SummaryVec
	HttpRequests *prometheus.CounterVec
	HttpErrors   *prometheus.CounterVec
}

var metricsMap Metrics
var customBuckets = []float64{0.1, 0.2, 0.5, 1, 2, 5, 10}

func registerMetrics(registry *prometheus.Registry) {
	metricsMap = Metrics{
		E2EDelay: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "http_e2e_delay",
				Help:       "E2E Delay in HTTP requests",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"service_from", "service_to"},
		),
		HttpErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_error_count",
				Help: "Number of http errors faced",
			},
			[]string{"service_from", "service_to"},
		),
		HttpRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_count",
				Help: "Number of http requests faced",
			},
			[]string{"service_from", "service_to"}),
	}
	registry.MustRegister(metricsMap.HttpErrors)
	registry.MustRegister(metricsMap.E2EDelay)
	registry.MustRegister(metricsMap.HttpRequests)
}

func PrometheusInit() {
	reg := prometheus.NewRegistry()
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	registerMetrics(reg)
}

func GetMetricsMap() Metrics {
	return metricsMap
}
