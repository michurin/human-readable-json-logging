package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func runSubprocessMode(lineFmt, errFmt func([]slogtotext.Pair) error) { //nolint:funlen
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

	deb("running subprocess: " + command + " " + strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = wr
	cmd.Stderr = wr

	err := cmd.Start()
	if err != nil {
		panic(err) // TODO
	}

	syncWait := make(chan error, 1)
	go func() {
		syncWait <- cmd.Wait()
	}()

	syncTerm := make(chan os.Signal, 1)
	signal.Notify(syncTerm, os.Interrupt, syscall.SIGTERM)

	syncSignal := make(chan os.Signal, 1)

	syncForceDone := make(chan struct{})

	exitCode := 0

LOOP:
	for {
		select {
		case err := <-syncWait: // have to break loop in this case
			deb("subprocess waiting is done")
			if err != nil {
				xerr := new(exec.ExitError)
				if errors.As(err, &xerr) {
					exitCode = xerr.ExitCode()
					if exitCode < 0 {
						exitCode = 1
					}
					break LOOP
				}
				panic(err) // TODO
			}
			break LOOP
		case sig := <-syncTerm:
			deb("pplog gets " + sig.String())
			syncSignal <- syscall.SIGINT
		case sig := <-syncSignal:
			deb("pplog sending " + sig.String() + " to subprocess")
			err := cmd.Process.Signal(sig)
			if err != nil {
				panic(err) // TODO
			}
			switch sig {
			case syscall.SIGINT:
				go func() {
					<-time.After(time.Second)
					syncSignal <- syscall.SIGTERM
				}()
			case syscall.SIGTERM:
				go func() {
					<-time.After(time.Second)
					syncSignal <- syscall.SIGKILL
				}()
			case syscall.SIGKILL:
				go func() {
					<-time.After(time.Second)
					close(syncForceDone)
				}()
			}
		case <-syncForceDone:
			panic("it seems the process can not be stopped")
		}
	}

	deb("stopping reader and waiting for it")
	_ = wr.Close() // TODO check error
	<-done

	os.Exit(exitCode)
}
