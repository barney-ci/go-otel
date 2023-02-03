package otel

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/samplers/jaegerremote"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	trace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type setupConfig struct {
	name            string
	envGate         bool
	shutdownTimeout time.Duration
	errPrinter      printer
	exporter        trace.SpanExporter
	sampler         trace.Sampler
	propagator      propagation.TextMapPropagator
}

type setupOptionFunc func(*setupConfig)

type printer func(string, ...interface{})

type closerFunc func() error

const EnvSamplerTemplateName = "OTEL_SAMPLER_JAEGER_CONFIG_URL_TEMPLATE"
const EnvGateName = "JAEGER_ENABLED"
const EnvGateCue = "true"

// nullExporter implements the trace.SpanExporter interface
type nullExporter struct{}

func (n nullExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	return nil
}

func (n nullExporter) Shutdown(ctx context.Context) error {
	return nil
}

// WithEnvGate causes a call to JaegerSetup to be a no-op
// unless the environment variable defined by EnvGatename
// is set to the value defined by EnvGateCue
func WithEnvGate() setupOptionFunc {
	return setupOptionFunc(func(opts *setupConfig) {
		opts.envGate = true
	})
}

// WithShutdownTimeout limits the amount of time
// that the close function returned by JaegerSetup may wait
func WithShutdownTimeout(t time.Duration) setupOptionFunc {
	return setupOptionFunc(func(opts *setupConfig) {
		opts.shutdownTimeout = t
	})
}

// WithGeneralPropagatorSetup causes JaegerSetup to configure
// the default propagator with some basic propagators
func WithGeneralPropagatorSetup() setupOptionFunc {
	p := propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
		UberTraceContext{},
	)
	return setupOptionFunc(func(opts *setupConfig) {
		opts.propagator = p
	})
}

// WithErrorPrinter causes JaegerSetup to call a printf-like
// function with error messages
func WithErrorPrinter(p printer) setupOptionFunc {
	return setupOptionFunc(func(opts *setupConfig) {
		opts.errPrinter = p
	})
}

// WithSampler causes JaegerSetup to configure Jaeger
// with the provided sampler only
func WithSampler(s trace.Sampler) setupOptionFunc {
	return setupOptionFunc(func(opts *setupConfig) {
		opts.sampler = s
	})
}

// WithRemoteSampler causes JaegerSetup to configure Jaeger
// with a remote sampler URL constructed using the environment
// variable defined by EnvSamplerTemplateName, falling back
// to any previously configured sampler
func WithRemoteSampler() setupOptionFunc {
	return setupOptionFunc(func(opts *setupConfig) {
		if samplerURL := os.Getenv(EnvSamplerTemplateName); samplerURL != "" {
			samplerURL = os.ExpandEnv(samplerURL)
			samplerURL = strings.ReplaceAll(samplerURL, "{}", url.QueryEscape(opts.name))
			opts.sampler = jaegerremote.New(opts.name,
				jaegerremote.WithSamplingServerURL(samplerURL),
				jaegerremote.WithInitialSampler(opts.sampler))
		}
	})
}

// WithAgentExpoter causes JaegerSetup to configure an
// exporter targeting the Jaeger agent endpoint
func WithAgentExporter() setupOptionFunc {
	return setupOptionFunc(func(opts *setupConfig) {
		exporter, err := jaeger.New(jaeger.WithAgentEndpoint())
		if err != nil {
			panic(fmt.Sprintf("cannot create jaeger exporter: %s", err))
		}
		opts.exporter = exporter
	})
}

func (p printer) printf(format string, args ...interface{}) {
	if p != nil {
		p(format, args...)
	}
}

// JaegerSetup returns a jaeger TracerProvider
// and a closer function to shut down the provider.
//
// Options order can be important. For example, WithRemoteSampler
// sets the sampler to one that falls back to any existing sampler,
// whilst WithSampler sets the sampler to the passed argument and
// overwrites the existing sampler.
//
// It's a good idea to pass WithErrorPrinter first, so errors
// raised by subsequent options will be sent to that callback.
func JaegerSetup(name string, with ...setupOptionFunc) (
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
		sampler:  trace.AlwaysSample(),
		exporter: nullExporter{},
	}
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
			opts.errPrinter.printf("%s", r)
		}
	}()
	for _, fn := range with {
		fn(opts)
	}

	if opts.envGate && os.Getenv(EnvGateName) != EnvGateCue {
		return
	}

	if opts.propagator != nil {
		otel.SetTextMapPropagator(opts.propagator)
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(opts.exporter),
		trace.WithSampler(opts.sampler),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name))),
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
		opts.errPrinter.printf("jaeger shutdown: %s", err)
		return err
	})

	return
}
