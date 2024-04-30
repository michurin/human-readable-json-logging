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
		require.Len(t, p, 2)
		require.Equal(t, "TEXT", p[0].K)
		require.Equal(t, "BINARY", p[1].K)
		out = append(out, fmt.Sprintf("%q (%q)", p[0].V, p[1].V))
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
		{name: "json", in: `{"a":1}`, exp: `a=1 RAW_INPUT={"a":1}`},
		{name: "empty", in: "", exp: ""},                                                                // has no effect
		{name: "nl", in: "\n\n", exp: `"" ("")|"" ("")`},                                                // invalid json
		{name: "invalid_json", in: `{"a":1}x`, exp: `"{\"a\":1}x" ("")`},                                // as is
		{name: "invalid_json_with_ctrl", in: `{"a":1}` + "\033" + `x`, exp: `"{\"a\":1}\x1bx" ("yes")`}, // as is with label binary=yes
		{name: "valid_but_too_long", in: `{"a":"123"}`, exp: `"{\"a\":\"123\"" ("")|"}" ("")`},
		{name: "valid_but_too_long_and_ok", in: `{"a":"123"}` + "\n" + `{"a":1}`, exp: `"{\"a\":\"123\"" ("")|"}" ("")|a=1 RAW_INPUT={"a":1}`},
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
