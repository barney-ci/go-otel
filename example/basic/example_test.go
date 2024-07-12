// Copyright (c) 2024 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package basic_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	otelb "barney.ci/go-otel"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func setup(ctx context.Context) {
	os.Setenv("JAEGER_ENABLED", "true")

	tp, closer, err := otelb.JaegerSetup(
		"barney.ci/go-otel/example/basic",
		// logger first so that any errors can be reported.
		otelb.WithLogger(logr.FromContextOrDiscard(ctx)),

		otelb.WithEnvGate(),
		otelb.WithShutdownTimeout(time.Minute),
		otelb.WithGeneralPropagatorSetup(),
		otelb.WithRemoteSampler(),
		otelb.WithAgentExporter(),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := closer(); err != nil {
			panic(fmt.Errorf("closer: %w", err))
		}
	}()

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("example")
}

func doTracing(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "span",
		trace.WithAttributes(attribute.String("colour", "red")),
	)
	defer span.End()

	val := 42
	span.AddEvent("log", trace.WithAttributes(attribute.Int("val", val)))
	span.RecordError(errors.New("found an error"))
	fmt.Println("sending a root span with an error")
}

func ExampleTracer() {
	// - OTEL_EXPORTER_JAEGER_AGENT_HOST is used for the agent address host
	// - OTEL_EXPORTER_JAEGER_AGENT_PORT is used for the agent address port

	ctx := context.TODO()
	setup(ctx)
	doTracing(ctx)
	// Output: sending a root span with an error
}
