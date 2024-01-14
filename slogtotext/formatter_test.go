package slogtotext_test

import (
	"bytes"
	"testing"

	"github.com/michurin/human-readable-json-logging/slogtotext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			template: "{{.A}}",
			in:       []slogtotext.Pair{{K: "A", V: "1"}},
			out:      "1",
		},
		{
			name:     "simple_not_value",
			template: "{{.A}}",
			in:       []slogtotext.Pair{},
			out:      "<no value>",
		},
		{
			name:     "time_formatter",
			template: `{{.A | tmf "2006-01-02T15:04:05Z07:00" "15:04:05"}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1975-12-02T12:01:02Z"}},
			out:      "12:01:02",
		},
		{
			name:     "time_formatter_invalid",
			template: `{{.A | tmf "2006-01-02" "2006-01-02"}}`,
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
			name:     "range_rm_pfx",
			template: `{{ range .ALL | rmByPfx "A" }}{{.K}}={{.V}};{{end}}`,
			in:       []slogtotext.Pair{{K: "A", V: "1"}, {K: "AA", V: "11"}, {K: "B", V: "2"}},
			out:      "B=2;",
		},
	} {
		cs := cs
		t.Run(cs.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			f := slogtotext.FormatterMust(buf, cs.template)
			f(cs.in)
			assert.Equal(t, cs.out, buf.String())
		})
	}
}

func TestFormatterInvalidTemplate(t *testing.T) {
	require.Panics(t, func() {
		slogtotext.FormatterMust(nil, "{{")
	})
}
