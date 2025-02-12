// Copyright (c) 2024 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package basic_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	otelb "barney.ci/go-otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func setup(ctx context.Context) {
	tp, closer, err := otelb.OtelSetup(
		ctx,
		"barney.ci/go-otel/example/basic",
		// logger first so that any errors can be reported.
		otelb.WithLogger(slog.Default()),

		otelb.WithEnvGate(),
		otelb.WithShutdownTimeout(time.Minute),
		otelb.WithGeneralPropagatorSetup(),
		otelb.WithRemoteSampler(),
		otelb.WithOtlpExporter(),
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
	err := errors.New("found an error")
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	fmt.Println("sending a root span with an error")
}

func ExampleTracer() {
	// - OTEL_EXPORTER_OTLP_ENDPOINT is used for the otlp exporter

	ctx := context.TODO()
	setup(ctx)
	doTracing(ctx)
	// Output: sending a root span with an error
}
