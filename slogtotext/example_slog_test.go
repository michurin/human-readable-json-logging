//go:build go1.21
// +build go1.21

package slogtotext_test

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func tweakNowToMakeTestReproducible(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{
			Key:   a.Key,
			Value: slog.TimeValue(time.Unix(186714000, 0).UTC()),
		}
	}
	return a
}

func Example_slog() {
	templateForJsonLogRecords := `{{ .time }} [{{ .level }}] {{ .msg }}{{ range .ALL | rm "time" "level" "msg" }} {{.K}}={{.V}}{{end}}` + "\n"
	templateForInvalidRecords := `INVALID JSON: {{ .text | printf "%q" }}` + "\n"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rd, wr, err := os.Pipe()
	panicIfError(err)

	defer func() {
		err = wr.Close() // The good idea is to close logging stream...
		panicIfError(err)
		<-ctx.Done() // ...and wait for all messages
	}()

	go func() {
		defer cancel()
		panicIfError(slogtotext.Read(
			rd,
			slogtotext.Formatter(os.Stdout, templateForJsonLogRecords),
			slogtotext.Formatter(os.Stdout, templateForInvalidRecords),
			1024))
	}()

	log := slog.New(slog.NewJSONHandler(wr, &slog.HandlerOptions{ReplaceAttr: tweakNowToMakeTestReproducible}))

	log.Info("Just log message")
	log.Error("Some error message", "customKey", "customValue")

	_, err = wr.Write([]byte("panic message\n")) // emulate wrong json in stream
	panicIfError(err)

	// output:
	// 1975-12-02T01:00:00Z [INFO] Just log message
	// 1975-12-02T01:00:00Z [ERROR] Some error message customKey=customValue
	// INVALID JSON: "panic message"
}
