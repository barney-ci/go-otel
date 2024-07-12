// Copyright (c) 2024 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package otel

import (
	"context"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func assertString(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("mismatch, expected %q, actual %q", expected, actual)
	}
}

func TestUberTraceIDExtraction(t *testing.T) {
	for _, tc := range []struct {
		header        string
		expectTraceID string
		expectSpanID  string
		expectFlag    byte
	}{
		// short trace id
		{
			header:        "5775daa08825b4b9:5c301b3cb0f66539:114d5e5bcc8bc4c8:1",
			expectTraceID: "00000000000000005775daa08825b4b9",
			expectSpanID:  "5c301b3cb0f66539",
			expectFlag:    1,
		},
		{
			header:        "5775daa08825b4b9:5c301b3cb0f66539:114d5e5bcc8bc4c8:01",
			expectTraceID: "00000000000000005775daa08825b4b9",
			expectSpanID:  "5c301b3cb0f66539",
			expectFlag:    1,
		},
		// long trace id
		{
			header:        "ee2ec3bb2402eb08625a76f762fb73bb:5c301b3cb0f66539:114d5e5bcc8bc4c8:1",
			expectTraceID: "ee2ec3bb2402eb08625a76f762fb73bb",
			expectSpanID:  "5c301b3cb0f66539",
			expectFlag:    1,
		},
		{
			header:        "ee2ec3bb2402eb08625a76f762fb73bb:5c301b3cb0f66539:114d5e5bcc8bc4c8:01",
			expectTraceID: "ee2ec3bb2402eb08625a76f762fb73bb",
			expectSpanID:  "5c301b3cb0f66539",
			expectFlag:    1,
		},
	} {
		t.Run(tc.header, func(t *testing.T) {
			u := &UberTraceContext{}
			h := http.Header{}
			h.Add("uber-trace-id", tc.header)
			ctx := u.Extract(context.Background(), propagation.HeaderCarrier(h))
			sc := trace.SpanContextFromContext(ctx)
			if !sc.IsValid() {
				t.Fatalf("span context was not valid")
			}
			assertString(t, tc.expectTraceID, sc.TraceID().String())
			assertString(t, tc.expectSpanID, sc.SpanID().String())
			if expected := trace.TraceFlags(tc.expectFlag); expected != sc.TraceFlags() {
				t.Fatalf("trace flags mismatch, expected: %v, actual: %v", expected, sc.TraceFlags())
			}
		})
	}
}
