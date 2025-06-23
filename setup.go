package otel

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/samplers/jaegerremote"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

type setupConfig struct {
	name            string
	envGate         bool
	shutdownTimeout time.Duration
	logger          *slog.Logger
	exporter        trace.SpanExporter
	sampler         trace.Sampler
	propagator      propagation.TextMapPropagator
}

type SetupOptionFunc func(*setupConfig)

type closerFunc func() error

var _ io.Closer = closerFunc(nil)

func (f closerFunc) Close() error {
	return f()
}

const EnvSamplingUrl = "OTEL_REMOTE_SAMPLING_URL"
const EnvGateCue = "true"

// These envs are standard. See:
// https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
// https://opentelemetry.io/docs/specs/otel/protocol/exporter/
const EnvGateName = "OTEL_SDK_DISABLED"

// nullExporter implements the trace.SpanExporter interface
type nullExporter struct{}

func (n nullExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	return nil
}

func (n nullExporter) Shutdown(ctx context.Context) error {
	return nil
}

// WithEnvGate causes a call to OtelSetup to be a no-op
// if the environment variable defined by EnvGatename
// is set to the value defined by EnvGateCue
func WithEnvGate() SetupOptionFunc {
	return SetupOptionFunc(func(opts *setupConfig) {
		opts.envGate = true
	})
}

// WithShutdownTimeout limits the amount of time
// that the close function returned by OtelSetup may wait
func WithShutdownTimeout(t time.Duration) SetupOptionFunc {
	return SetupOptionFunc(func(opts *setupConfig) {
		opts.shutdownTimeout = t
	})
}

// WithGeneralPropagatorSetup causes OtelSetup to configure
// the default propagator with some basic propagators
func WithGeneralPropagatorSetup() SetupOptionFunc {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
		UberTraceContext{},
	)
	return SetupOptionFunc(func(opts *setupConfig) {
		opts.propagator = p
	})
}

// WithLogger configures the given logger to be used for printing errors
// or info at runtime emitted by the tracer implementation. If unset,
// a default value of slog.Default() will be used.
func WithLogger(logger *slog.Logger) SetupOptionFunc {
	return SetupOptionFunc(func(opts *setupConfig) {
		opts.logger = logger
	})
}

// WithSampler causes OtelSetup to configure otel
// with the provided sampler only
func WithSampler(s trace.Sampler) SetupOptionFunc {
	return SetupOptionFunc(func(opts *setupConfig) {
		opts.sampler = s
	})
}

// WithOtlpExporter causes OtelSetup to configure an
// exporter targeting the exporter otlp endpoint
func WithOtlpExporter() SetupOptionFunc {
	return SetupOptionFunc(func(opts *setupConfig) {
		exporter, err := otlptracegrpc.New(context.Background())
		if err != nil {
			panic(fmt.Sprintf("cannot create otlp exporter: %s", err))
		}

		opts.exporter = exporter
	})
}

// WithRemoteSampler causes OtelSetup to be configured
// with a remote sampler URL constructed using the environment
// variable defined by EnvSamplingUrl, falling back
// to any previously configured sampler
func WithRemoteSampler() SetupOptionFunc {
	return SetupOptionFunc(func(opts *setupConfig) {
		if samplingURL := os.Getenv(EnvSamplingUrl); samplingURL != "" {
			if strings.Contains(samplingURL, "{}") {
				panic(fmt.Sprintf("%s no longer supports {} macro; "+
					"please see the barney.ci/go-otel readme", EnvSamplingUrl))
			}
			samplingURL = os.ExpandEnv(samplingURL)
			opts.sampler = jaegerremote.New(opts.name,
				jaegerremote.WithSamplingServerURL(samplingURL),
				jaegerremote.WithInitialSampler(opts.sampler),
				jaegerremote.WithLogger(logr.FromSlogHandler(opts.logger.Handler())),
			)
		}
	})
}

func getIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IP address found")
}

// OtelSetup returns a otel TracerProvider
// and a closer function to shut down the provider.
//
// Options order can be important. For example, WithRemoteSampler
// sets the sampler to one that falls back to any existing sampler,
// whilst WithSampler sets the sampler to the passed argument and
// overwrites the existing sampler.
//
// It's a good idea to pass WithLogger first, so errors
// raised by subsequent options will be sent to that callback.
func OtelSetup(ctx context.Context, name string, with ...SetupOptionFunc) (
	tp *trace.TracerProvider, closer closerFunc, err error,
) {
	// Always return working no-ops instead of nils
	defer func() {
		if tp == nil {
			tp = trace.NewTracerProvider()
		}
		if closer == nil {
			closer = closerFunc(func() error { return nil })
		}
	}()

	// Apply options and return an error if one panics
	opts := &setupConfig{
		name:     name,
		sampler:  trace.ParentBased(trace.AlwaysSample()),
		exporter: nullExporter{},
		logger:   slog.Default(),
	}
	defer func() {
		if r := recover(); r != nil {
			opts.logger.ErrorContext(ctx, "panic occurred in OtelSetup", "error", r)
		}
	}()
	for _, fn := range with {
		fn(opts)
	}

	if opts.envGate && os.Getenv(EnvGateName) == EnvGateCue {
		return
	}

	if opts.propagator != nil {
		otel.SetTextMapPropagator(opts.propagator)
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(name),
		semconv.TelemetrySDKNameKey.String("opentelemetry"),
		semconv.TelemetrySDKVersionKey.String(otel.Version()),
	}

	if ip, err := getIPAddress(); err != nil {
		opts.logger.ErrorContext(ctx, "failed to find host IP address", "error", err)
	} else {
		attrs = append(attrs, semconv.HostIPKey.String(ip))
	}

	if host, err := os.Hostname(); err != nil {
		opts.logger.ErrorContext(ctx, "os.Hostname() failed", "error", err)
	} else {
		attrs = append(attrs, semconv.HostNameKey.String(host))
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(opts.exporter),
		trace.WithSampler(opts.sampler),
		trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)),
	)

	closer = closerFunc(func() error {
		var ctx context.Context
		var cancel context.CancelFunc
		if opts.shutdownTimeout > 0 {
			ctx, cancel = context.WithTimeout(
				context.Background(), opts.shutdownTimeout)
			defer cancel()
		} else {
			ctx = context.Background()
		}
		err := tp.Shutdown(ctx)
		if err != nil {
			opts.logger.ErrorContext(ctx, "otel shutdown error", "error", err)
		}
		return err
	})

	return
}
