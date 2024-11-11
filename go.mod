module barney.ci/go-otel

go 1.22

toolchain go1.23.1

require (
	github.com/go-logr/logr v1.4.2
	github.com/go-logr/stdr v1.2.2
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.26.0
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)
