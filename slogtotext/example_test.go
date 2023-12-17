package slogtotext_test

import (
	"os"
	"strings"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func Example() {
	f := slogtotext.Formatter(os.Stdout, `x={{.x}}{{if .UNKNOWN}} UNKNOWN:{{range .UNKNOWN | remove "p" "q"}} {{.K}}={{.V}}{{end}}{{end}}`)
	g := slogtotext.Formatter(os.Stdout, `INVALID LINE: {{ .text | printf "%q" }}`)
	buf := strings.NewReader(`
{"x": 100}
{"x": 1, "y": {
	"a": 2,
	"b": 3
}, "p": 9, "q": 9}
here`)
	err := slogtotext.Read(buf, f, g, 1024)
	if err != nil {
		panic(err)
	}
	// output:
	// x=100
	// x=1 UNKNOWN: y.a=2 y.b=3
	// INVALID LINE: "here"
}
