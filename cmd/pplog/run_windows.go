//go:build windows
// +build windows

package main

import (
	"context"
	"flag"
	"io"
	"os"
	"os/exec"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func runSubprocessMode() {
	deb("run subprocess mode (windows)")

	ctx := context.Background()

	rd, wr := io.Pipe()

	f, g := prepareFormatters()
	go func() {
		err := slogtotext.Read(rd, f, g, 32768)
		if err != nil {
			deb("reader is finished with err: " + err.Error())
			return
		}
		deb("reader is finished")
	}()

	args := flag.Args()[1:]
	command := flag.Args()[0]

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = wr
	cmd.Stderr = wr

	err := cmd.Run()
	deb("subprocess is finished")
	if err != nil {
		printError(err)
	}
	exitCode := cmd.ProcessState.ExitCode() // exit code can be set even if error
	if exitCode < 0 {
		exitCode = 1
	}
	os.Exit(exitCode)
}
