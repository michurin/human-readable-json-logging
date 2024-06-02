//go:build windows
// +build windows

package main

import (
	"io"
	"os"
	"os/exec"
)

func catchSigChild(rd *io.PipeReader) {
	// do nothing, windows does not support SIGCHLD
}

func signalToCmd(cmd *exec.Cmd, sig os.Signal) {
	if sig == os.Interrupt {
		// cmd.Process.Signal(os.Interrupt) returns an error on windows
		// just skip, windows itself sends INT to all processes attached to console
		return
	}
	err := cmd.Process.Signal(sig)
	if err != nil {
		printError(err)
	}
}
