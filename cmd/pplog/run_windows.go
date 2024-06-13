//go:build windows
// +build windows

package main

import (
	"flag"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func runSubprocessMode() {
	deb("run subprocess mode (windows)")

	rd, wr := io.Pipe()

	f, g := prepareFormatters()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := slogtotext.Read(rd, f, g, 32768)
		if err != nil {
			deb("reader is finished with err: " + err.Error())
			return
		}
		deb("reader is finished")
	}()

	args := flag.Args()[1:]
	command := flag.Args()[0]

	cmd := exec.Command(command, args...)
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

	wg.Wait() // allow reader to process records
	os.Exit(exitCode)
}
