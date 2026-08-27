package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/samirmaji-tech/gke-microservice-app/internal/handlers"
	"github.com/samirmaji-tech/gke-microservice-app/internal/metrics"
	appmw "github.com/samirmaji-tech/gke-microservice-app/internal/middleware"
	"github.com/samirmaji-tech/gke-microservice-app/internal/telemetry"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	// Structured JSON logs to stdout -> collected by GKE logging / Loki.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics.MustRegister()

	// Tracing is optional; enabled only when the collector endpoint is set.
	var shutdownTracer func(context.Context) error
	if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		sd, err := telemetry.InitTracer(ctx, getenv("OTEL_SERVICE_NAME", "gke-microservice"), ep)
		if err != nil {
			logger.Error("failed to init tracer", slog.Any("err", err))
		} else {
			shutdownTracer = sd
			logger.Info("tracing enabled", slog.String("endpoint", ep))
		}
	}

	health := handlers.NewHealth()
	api := handlers.NewAPI()

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(func(next http.Handler) http.Handler { return appmw.Observability(logger, next) })

	// Probes + metrics: kept out of tracing to avoid noise.
	r.Get("/healthz", health.Livez)
	r.Get("/readyz", health.Readyz)
	r.Handle("/metrics", promhttp.Handler())

	// Business routes wrapped with OpenTelemetry server instrumentation.
	r.Group(func(gr chi.Router) {
		gr.Use(func(next http.Handler) http.Handler { return otelhttp.NewHandler(next, "http.server") })
		gr.Get("/", api.Root)
		gr.Get("/api/v1/items/{id}", api.GetItem)
	})

	srv := &http.Server{
		Addr:              getenv("HTTP_ADDR", ":8080"),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Simulate warmup, then flip readiness so /readyz returns 200.
	go func() {
		time.Sleep(2 * time.Second)
		health.SetReady(true)
		logger.Info("service ready")
	}()

	go func() {
		logger.Info("server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", slog.Any("err", err))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Fail readiness first so the LB drains us before we stop.
	health.SetReady(false)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("err", err))
	}
	if shutdownTracer != nil {
		_ = shutdownTracer(shutdownCtx)
	}
	logger.Info("stopped cleanly")
}
