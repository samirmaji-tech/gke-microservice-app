package metrics

import "github.com/prometheus/client_golang/prometheus"

// Custom RED-style metrics exposed at /metrics for Prometheus to scrape.
var (
	// Rate + Errors: total requests labelled by outcome.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed, by method, route and status.",
		},
		[]string{"method", "route", "status"},
	)

	// Duration: request latency histogram (feeds Grafana p50/p95/p99).
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	// Saturation-ish: current concurrent in-flight requests.
	InFlightRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests.",
		},
	)
)

// MustRegister wires the collectors into the default registry.
func MustRegister() {
	prometheus.MustRegister(HTTPRequestsTotal, HTTPRequestDuration, InFlightRequests)
}
