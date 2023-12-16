//go:build go1.21
// +build go1.21

package slogtotext_test

import (
	"log/slog"
	"os"
	"time"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

//nolint:errcheck
func ExamplePPLog_slog() {
	w := slogtotext.PPLog(os.Stdout, "", "", nil, nil, 0)
	log := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: false,
		Level:     nil,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey { // tweak now() to make test reproducible
				return slog.Attr{
					Key:   a.Key,
					Value: slog.TimeValue(time.Unix(186714000, 0).UTC()),
				}
			}
			return a
		},
	}))
	log.Info("Just log message")
	log.Error("Some error message", "customKey", "customValue")
	// output:
	// 1975-12-02T01:00:00Z [INFO] Just log message
	// 1975-12-02T01:00:00Z [ERROR] Some error message customKey=customValue
}
