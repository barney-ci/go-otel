module barney.ci/go-otel

go 1.16

require (
	github.com/go-logr/logr v1.4.1
	github.com/go-logr/stdr v1.2.2
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.6.0
	go.opentelemetry.io/otel v1.22.0
	go.opentelemetry.io/otel/exporters/jaeger v1.11.2
	go.opentelemetry.io/otel/sdk v1.22.0
	go.opentelemetry.io/otel/trace v1.22.0
)
