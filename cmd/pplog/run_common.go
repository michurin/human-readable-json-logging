package main

import (
	"flag"
	"io"
	"os"
	"os/exec"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func runSubprocessMode(lineFmt, errFmt func([]slogtotext.Pair) error) {
	deb("run subprocess mode")

	rd, wr := io.Pipe()

	done := make(chan struct{})

	go func() {
		err := slogtotext.Read(rd, lineFmt, errFmt, buffSize)
		if err != nil {
			deb("reader is finished with err: " + err.Error())
			return
		}
		deb("reader is finished")
		close(done)
	}()

	args := flag.Args()[1:]
	command := flag.Args()[0]

	cmd := exec.Command(command, args...)
	cmd.Stdout = wr
	cmd.Stderr = wr

	err := cmd.Run()
	_ = rd.Close() // we have to do it to finalize reader
	deb("subprocess is finished")
	if err != nil {
		printError(err)
	}
	exitCode := cmd.ProcessState.ExitCode() // exit code can be set even if error
	if exitCode < 0 {
		exitCode = 1
	}

	<-done

	os.Exit(exitCode)
}
