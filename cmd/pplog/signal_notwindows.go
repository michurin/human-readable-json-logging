//go:build !windows
// +build !windows

package main

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func catchSigChild(rd *io.PipeReader) {
	catchSigChld := make(chan os.Signal, 1)
	signal.Notify(catchSigChld, syscall.SIGCHLD)
	go func() {
		sig := <-catchSigChld
		deb("catch signal: " + sig.String())
		time.Sleep(time.Second) // we give a second to collect the last data; this signal obtaining from group, nor from direct child
		rd.Close()
	}()
}

func signalToCmd(cmd *exec.Cmd, sig os.Signal) {
	err := cmd.Process.Signal(sig)
	if err != nil {
		printError(err)
	}
}
