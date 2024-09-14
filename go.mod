module barney.ci/go-otel

go 1.22

toolchain go1.23.1

require (
	github.com/go-logr/logr v1.4.2
	github.com/go-logr/stdr v1.2.2
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.24.0
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.30.0
	go.opentelemetry.io/otel/trace v1.30.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)
