package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/example/gke-microservice-app/internal/metrics"
)

// statusRecorder captures the status code written by downstream handlers.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Observability records Prometheus RED metrics AND a structured JSON log line
// for every request. Uses the chi route PATTERN (e.g. /api/v1/items/{id})
// as the metric label to avoid unbounded label cardinality.
func Observability(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.InFlightRequests.Inc()
		defer metrics.InFlightRequests.Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		dur := time.Since(start)
		route := routePattern(r)

		metrics.HTTPRequestsTotal.
			WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		metrics.HTTPRequestDuration.
			WithLabelValues(r.Method, route).Observe(dur.Seconds())

		logger.Info("http_request",
			slog.String("method", r.Method),
			slog.String("route", route),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Float64("duration_ms", float64(dur.Microseconds())/1000.0),
			slog.String("remote", r.RemoteAddr),
		)
	})
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if p := rctx.RoutePattern(); p != "" {
			return p
		}
	}
	return r.URL.Path
}
