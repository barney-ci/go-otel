# barney.ci/go-otel

This repo contains OpenTelemetry-related code.

### OtelSetup

OtelSetup provides boilerplate, slightly Arista-optimized production of Otel TracerProviders.

Typical applications require some environment. Kubernetes example:

```
kind: Deployment
spec:
  template:
    spec:
      containers:
        env:
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://$(HOST_IP):4317"
        - name: OTEL_REMOTE_SAMPLING_URL
          value: "http://$(HOST_IP):5778/sampling"
```

The old envs `OTEL_EXPORTER_JAEGER_AGENT_HOST`, `OTEL_SAMPLER_JAEGER_CONFIG_URL_TEMPLATE`, and `JAEGER_ENABLED` are no longer used.
The grpc otel exporter can be configured through the standard envs described https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc#pkg-overview

### EnvironCarrier

EnvironCarrier provides a TextMapCarrier interface to the process environment.

### UberTraceContext

UberTraceContext is a propagator that supports the "Jaeger native propagation format", better known as the "uber-trace-id" header.
