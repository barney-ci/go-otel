// Copyright (c) 2022 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package otel

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// EnvironCarrier provides a TextMapCarrier interface to the process environment.
type EnvironCarrier struct{}

// EnvironCarrierPrefix defines the prefix we attach to
// carrier key names to store them in the process environment.
const EnvironCarrierPrefix = "OTELTEXTMAP_" // exactly 1 underscore, trailing

// ContextWithEnvPropagation returns a copy of a parent Context
// with trace context propagation from the process environment.
func ContextWithEnvPropagation(parent context.Context) context.Context {
	return otel.GetTextMapPropagator().Extract(parent, EnvironCarrier{})
}

// MapCarrierAsEnviron returns the contents of a MapCarrier as
// a slice of "key=value" strings, suitable for e.g. os.exec.Cmd.Env
func MapCarrierAsEnviron(mc propagation.MapCarrier) []string {
	ret := make([]string, len(mc))
	i := 0
	for k, v := range mc {
		ret[i] = fmt.Sprintf("%s%s=%s", EnvironCarrierPrefix, k, v)
		i++
	}
	return ret
}

func transcode(s string, specialIn, specialOut byte) (string, error) {
	var ret []byte
	var out byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == specialIn:
			out = specialOut
		case c >= 'A' && c <= 'Z':
			out = c
		case c >= 'a' && c <= 'z':
			out = c
		default:
			return "", fmt.Errorf("cannot handle character '%c'", c)
		}
		ret = append(ret, out)
	}
	return string(bytes.ToUpper(ret)), nil
}

func envEncode(s string) (string, error) {
	return transcode(s, '-', '_')
}

func envDecode(s string) (string, error) {
	return transcode(s, '_', '-')
}

func (e EnvironCarrier) Get(key string) string {
	envKey, err := envEncode(key)
	if err != nil {
		panic(err)
	}
	envName := EnvironCarrierPrefix + envKey
	return os.Getenv(envName)
}

func (e EnvironCarrier) Set(key string, value string) {
	envKey, err := envEncode(key)
	if err != nil {
		panic(err)
	}
	envName := EnvironCarrierPrefix + envKey
	os.Setenv(envName, value)
}

func (e EnvironCarrier) Keys() []string {
	ret := make([]string, 0)
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, EnvironCarrierPrefix) {
			continue
		}
		envName := strings.SplitN(env, "=", 2)[0]
		varName := strings.SplitN(envName, "_", 2)[1]
		if varName == "" {
			continue
		}
		key, err := envDecode(varName)
		if err != nil {
			continue
		}
		ret = append(ret, key)
	}
	return ret
}
