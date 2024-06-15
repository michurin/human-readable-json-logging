//go:build !windows
// +build !windows

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func runSubprocessModeChild() {
	deb("run child mode")

	target := flag.Args()
	binary, err := exec.LookPath(target[0])
	if err != nil {
		panic(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	chFiles := make([]uintptr, 3) //nolint:mnd // in, out, err
	chFiles[0] = r.Fd()
	chFiles[1] = os.Stdout.Fd()
	chFiles[2] = os.Stderr.Fd()

	selfBinaty, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pid, err := syscall.ForkExec(selfBinaty, os.Args[:1], &syscall.ProcAttr{
		Dir:   cwd,
		Env:   os.Environ(),
		Files: chFiles,
		Sys:   nil,
	})
	if err != nil {
		panic(selfBinaty + ": " + err.Error())
	}
	deb(fmt.Sprintf("subprocess pid: %d", pid))

	err = syscall.Dup2(int(w.Fd()), syscall.Stdout) // os.Stdout = w
	if err != nil {
		panic(err)
	}
	err = syscall.Dup2(int(w.Fd()), syscall.Stderr) // os.Stderr = w
	if err != nil {
		panic(err)
	}

	err = syscall.Exec(binary, target, os.Environ())
	if err != nil {
		panic(binary + ": " + err.Error())
	}
}
