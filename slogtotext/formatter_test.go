//nolint:funlen // it's ok for tests
package slogtotext_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func TestFormatter(t *testing.T) {
	for _, cs := range []struct {
		name     string
		template string
		in       []slogtotext.Pair
		out      string
	}{
		{
			name:     "nil",
			template: "OK",
			in:       nil,
			out:      "OK",
		},
		{
			name:     "simple",
			template: "{{ .A }}",
			in:       []slogtotext.Pair{{K: "A", V: "1"}},
			out:      "1",
		},
		{
			name:     "simple_not_value",
			template: "{{ .A }}",
			in:       []slogtotext.Pair{},
			out:      "<no value>",
		},
		{
			name:     "time_formatter",
			template: `{{ .A | tmf "2006-01-02T15:04:05Z07:00" "15:04:05" }}`,
			in:       []slogtotext.Pair{{K: "A", V: "1975-12-02T12:01:02Z"}},
			out:      "12:01:02",
		},
		{
			name:     "time_formatter_invalid",
			template: `{{ .A | tmf "2006-01-02" "2006-01-02" }}`,
			in:       []slogtotext.Pair{{K: "A", V: "1975-xii-02"}},
			out:      `parsing time "1975-xii-02" as "2006-01-02": cannot parse "xii-02" as "01"`,
		},
		{
			name:     "range",
			template: `{{ range .ALL }}{{.K}}={{.V}};{{end}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1"}, {K: "AA", V: "11"}, {K: "B", V: "2"}},
			out:      "A=1;AA=11;B=2;",
		},
		{
			name:     "range_rm",
			template: `{{ range .ALL | rm "A" }}{{.K}}={{.V}};{{end}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1"}, {K: "AA", V: "11"}, {K: "B", V: "2"}},
			out:      "AA=11;B=2;",
		},
		{
			name:     "range_rm_multi",
			template: `{{ range .ALL | rm "A" "B" }}{{.K}}={{.V}};{{end}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1"}, {K: "AA", V: "11"}, {K: "B", V: "2"}},
			out:      "AA=11;",
		},
		{
			name:     "range_rm_pfx",
			template: `{{ range .ALL | rmByPfx "A" }}{{.K}}={{.V}};{{end}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1"}, {K: "AA", V: "11"}, {K: "B", V: "2"}},
			out:      "B=2;",
		},
		{
			name:     "range_rm_pfx_multi",
			template: `{{ range .ALL | rmByPfx "A" "B" }}{{.K}}={{.V}};{{end}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1"}, {K: "AA", V: "11"}, {K: "B", V: "2"}, {K: "BB", V: "22"}, {K: "C", V: "3"}},
			out:      "C=3;",
		},
		{
			name:     "trim_space",
			template: `{{ .A | trimSpace }}`,
			in:       []slogtotext.Pair{{K: "A", V: " X "}},
			out:      `X`,
		},
		{
			name:     "sprig_function", // just to be sure that https://masterminds.github.io/sprig/ are on
			template: `{{ upper "hello" }}`,
			in:       nil,
			out:      "HELLO",
		},
	} {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			f := slogtotext.MustFormatter(buf, cs.template)
			err := f(cs.in)
			require.NoError(t, err)
			assert.Equal(t, cs.out, buf.String())
		})
	}
}

func TestFormatter_errors(t *testing.T) {
	for _, cs := range []struct {
		name     string
		template string
		err      string
	}{
		{
			name:     "range_rm_wrong_type",
			template: `{{ range .ALL | rm "A" true }}{{.K}}={{.V}};{{end}}`,
			err:      `template: base:1:16: executing "base" at <rm "A" true>: error calling rm: Invalid type: idx=1: bool: true`,
		},
		{
			name:     "range_rm_wrong_input_type",
			template: `{{ range 1 | rm "A" }}{{.K}}={{.V}};{{end}}`,
			err:      `template: base:1:13: executing "base" at <rm "A">: error calling rm: Invalid type: int: 1: only .ALL allows`,
		},
		{
			name:     "range_rm_noargs",
			template: `{{ range .ALL | rm }}{{.K}}={{.V}};{{end}}`,
			err:      `template: base:1:16: executing "base" at <rm>: error calling rm: Invalid number of args: 1: [[]]`,
		},
		{
			name:     "range_rm_pfx_wrong_type",
			template: `{{ range .ALL | rmByPfx "A" true }}{{.K}}={{.V}};{{end}}`,
			err:      `template: base:1:16: executing "base" at <rmByPfx "A" true>: error calling rmByPfx: Invalid type: idx=1: bool: true`,
		},
		{
			name:     "range_rm_pfx_wrong_input_type",
			template: `{{ range 1 | rmByPfx "A" }}{{.K}}={{.V}};{{end}}`,
			err:      `template: base:1:13: executing "base" at <rmByPfx "A">: error calling rmByPfx: Invalid type: int: 1: only .ALL allows`,
		},
		{
			name:     "range_rm_pfx_noargs",
			template: `{{ range .ALL | rmByPfx }}{{.K}}={{.V}};{{end}}`,
			err:      `template: base:1:16: executing "base" at <rmByPfx>: error calling rmByPfx: Invalid number of args: 1: [[]]`,
		},
	} {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			f := slogtotext.MustFormatter(buf, cs.template)
			err := f([]slogtotext.Pair{})
			require.EqualError(t, err, cs.err)
			assert.Empty(t, buf.String())
		})
	}
}

func TestFormatter_invalidArgs(t *testing.T) {
	for _, cs := range []struct {
		name     string
		template string
		out      string
	}{
		{
			name:     "wrong_time",
			template: `{{ 1 | tmf "2006-01-02" "2006-01-02" }}`,
			out:      `invalid time type: int (1)`,
		},
		{
			name:     "wrong_string",
			template: `{{ 1 | trimSpace " ok " }}`,
			out:      `ok 1`,
		},
	} {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			f := slogtotext.MustFormatter(buf, cs.template)
			err := f([]slogtotext.Pair{})
			require.NoError(t, err)
			assert.Equal(t, cs.out, buf.String())
		})
	}
}

func TestMustFormatter_invalid(t *testing.T) {
	for _, cs := range []struct {
		name     string
		template string
		value    string
	}{
		{
			name:     "function",
			template: "{{ . | notExists }}",
			value:    `template: base:1: function "notExists" not defined`,
		},
		{
			name:     "template",
			template: "{{",
			value:    "template: base:1: unclosed action",
		},
	} {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			require.PanicsWithValue(t, cs.value, func() {
				slogtotext.MustFormatter(nil, cs.template)
			})
		})
	}
}
