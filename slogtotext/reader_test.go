package slogtotext_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func pairs(p []slogtotext.Pair) string {
	x := make([]string, len(p))
	for i, v := range p {
		x[i] = v.K + "=" + v.V
	}
	return strings.Join(x, "\x20")
}

func collector(t *testing.T) (
	func([]slogtotext.Pair) error,
	func([]slogtotext.Pair) error,
	func() string,
) {
	t.Helper()
	out := []string(nil)
	f := func(p []slogtotext.Pair) error {
		out = append(out, pairs(p))
		return nil
	}
	g := func(p []slogtotext.Pair) error {
		t.Helper()
		require.Len(t, p, 1)
		require.Equal(t, "text", p[0].K)
		out = append(out, fmt.Sprintf("%q", p[0].V))
		return nil
	}
	c := func() string {
		return strings.Join(out, "|")
	}
	return f, g, c
}

func TestReader(t *testing.T) {
	t.Parallel()
	for _, cs := range []struct {
		name string
		in   string
		exp  string
	}{
		{name: "empty", in: "", exp: ""},
		{name: "nl", in: "\n\n", exp: `""|""`},
		{name: "json", in: `{"a":1}`, exp: "a=1"},
		{name: "invalid_json", in: `{"a":1}x`, exp: `"{\"a\":1}x"`}, // as is
	} {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			buf := strings.NewReader(cs.in)
			f, g, c := collector(t)
			err := slogtotext.Read(buf, f, g, 10)
			require.NoError(t, err)
			assert.Equal(t, cs.exp, c())
		})
	}
}
