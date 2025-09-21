//go:build windows || (linux && (arm64 || loong64 || riscv64))

package main

import (
	"fmt"
	"os"
)

func runSubprocessModeChild() {
	fmt.Fprintln(os.Stderr, "Child mode is not supported in MS Windows")
	os.Exit(1)
}
