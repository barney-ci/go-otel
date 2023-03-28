package otel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranscode(t *testing.T) {
	var s string
	var err error

	_, err = envEncode("hello_there")
	require.Error(t, err)

	_, err = envDecode("hello-there")
	require.Error(t, err)

	s, err = envEncode("hello-there")
	require.NoError(t, err)
	require.Equal(t, s, "HELLO_THERE")

	s, err = envDecode("HELLO_THERE")
	require.NoError(t, err)
	require.Equal(t, s, "HELLO-THERE")
}

const (
	theUber   = "uber-trace-id"
	theSponge = "sPoNgEbOb"
)

func TestCarrier(t *testing.T) {
	var s string

	e := EnvironCarrier{}

	e.Set(theUber, "12345")
	e.Set(theSponge, "xyzzy")

	k := e.Keys()
	require.Contains(t, k, "UBER-TRACE-ID")
	require.Contains(t, k, "SPONGEBOB")
	require.NotContains(t, k, "SPIDERMAN")

	s = e.Get(theUber)
	require.Equal(t, s, "12345")

	s = e.Get(theSponge)
	require.Equal(t, s, "xyzzy")
}
