# barney.ci/go-otel

This repo contains OpenTelemetry-related code.

### JaegerSetup

JaegerSetup provides boilerplate, slightly Arista-optimized production of Jaeger TracerProviders.

Typical applications require some environment. Kubernetes example:

```
kind: Deployment
spec:
  template:
    spec:
      containers:
        env:
        - name: JAEGER_ENABLED
          value: "true"
        - name: OTEL_EXPORTER_JAEGER_AGENT_HOST
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: OTEL_SAMPLER_JAEGER_CONFIG_URL_TEMPLATE
          value: http://${OTEL_EXPORTER_JAEGER_AGENT_HOST}:5778/
```

Please note that there are a few similar-looking URLs floating around for the URL_TEMPLATE.
The different URLs are meant for different collector libraries and have subtly different
output. You likely need to change your URL_TEMPLATE if you are migrating from a different
telemetry library so note carefully this syntax.

### EnvironCarrier

EnvironCarrier provides a TextMapCarrier interface to the process environment.

### UberTraceContext

UberTraceContext is a propagator that supports the "Jaeger native propagation format", better known as the "uber-trace-id" header.
