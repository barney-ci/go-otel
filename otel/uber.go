// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

// Some portions derived from the TraceContext propagator
// Copyright The OpenTelemetry Authors (Apache License, Version 2.0)

package otel

import (
	"context"
	"encoding/hex"
	"fmt"
	"regexp"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	uberHeader = "uber-trace-id"
)

// UberTraceContext is a propagator that supports the "Jaeger native propagation
// format", better known as the "uber-trace-id" header. See:
// https://www.jaegertracing.io/docs/1.40/client-libraries/#propagation-format
//
// UberTraceContext will propagate the uber-trace-id header to guarantee traces
// employing this type of header are not broken. It is up to the users of this
// propagator to choose if they want to participate in a trace by modifying the
// uber-trace-id header and relevant parts of the uber-trace-id header
// containing their proprietary information.
//
// UberTraceContext operates on the came principle as the upstream TraceContext
// propagator, which injects and extracts the "W3C trace context format", better
// known as the "traceparent" header.
//
// When a CompositeTextMapPropagator combines TraceContext and UberTraceContext
// propagators, SpanContexts will be propagated forward as both types of header,
// and both inbound header types will be extractable into a local SpanContext
// (with the later-defined propagator's header taking overriding precedence).
type UberTraceContext struct{}

var _ propagation.TextMapPropagator = UberTraceContext{}
var uberHdrRegExp = regexp.MustCompile("^(?P<traceID>[0-9a-f]{1,32}):(?P<spanID>[0-9a-f]{1,16}):(?P<parentSpanID>[0-9a-f]{1,16}):(?P<traceFlags>[0-9a-f]{1,2})(?:-.*)?$")

// Inject sets uber-trace-id from the Context into the carrier.
func (tc UberTraceContext) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}

	// Clear all flags other than the trace-context supported sampling bit.
	// The following flags would otherwise be valid for Jaeger:
	//     flagSampled  = 1
	//     flagDebug    = 2
	//     flagFirehose = 8
	flags := sc.TraceFlags() & trace.FlagsSampled

	h := fmt.Sprintf("%s:%s:%016x:%s",
		sc.TraceID(), // trace-id
		sc.SpanID(),  // span-id
		0,            // parent-span-id (deprecated in spec)
		flags,        // flags
	)
	carrier.Set(uberHeader, h)
}

// Extract reads uber-trace-id from the carrier into a returned Context.
//
// The returned Context will be a copy of ctx and contain the extracted
// uber-trace-id as the remote SpanContext. If the extracted uber-trace-id
// is invalid, the passed ctx will be returned directly instead.
func (tc UberTraceContext) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc := tc.extract(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (tc UberTraceContext) extract(carrier propagation.TextMapCarrier) trace.SpanContext {
	h := carrier.Get(uberHeader)
	if h == "" {
		return trace.SpanContext{}
	}

	matches := uberHdrRegExp.FindStringSubmatch(h)

	if len(matches) < 5 { // four subgroups plus the overall match
		return trace.SpanContext{}
	}

	var scc trace.SpanContextConfig

	traceID, err := decodeHexID(matches[1], 16) // 128 bits
	if err != nil {
		return trace.SpanContext{}
	}
	copy(scc.TraceID[:], traceID[:16])

	spanID, err := decodeHexID(matches[2], 8) // 64 bits
	if err != nil {
		return trace.SpanContext{}
	}
	copy(scc.SpanID[:], spanID[:8])

	flags, err := hex.DecodeString(matches[4])
	if err != nil {
		return trace.SpanContext{}
	}
	// Clear all flags other than the trace-context supported sampling bit.
	scc.TraceFlags = trace.TraceFlags(flags[0]) & trace.FlagsSampled

	sc := trace.NewSpanContext(scc)
	if !sc.IsValid() {
		return trace.SpanContext{}
	}

	return sc
}

// Fields returns the keys whose values are set with Inject.
func (tc UberTraceContext) Fields() []string {
	return []string{uberHeader}
}

// The spec states that receivers must accept trace and span ID's
// shorter than the expected field size, and zero-pad them on the left;
// and states that a zero value is not a valid trace or span ID.
func decodeHexID(s string, size int) ([]byte, error) {
	var allZeroes bool = true

	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if size < len(data) {
		return nil, fmt.Errorf("string too large for size")
	}
	pad := size - len(data)
	ret := make([]byte, size, size)
	for i, x := range data {
		ret[i+pad] = x
		if x != 0 {
			allZeroes = false
		}
	}
	if allZeroes {
		return nil, fmt.Errorf("zero value is invalid")
	}
	return ret, nil
}
