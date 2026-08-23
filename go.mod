module github.com/example/gke-microservice-app

go 1.22

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/prometheus/client_golang v1.19.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
)
