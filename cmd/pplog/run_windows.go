//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
)

func runSubprocessModeChild() {
	fmt.Fprintln(os.Stderr, "Child mode is not supported in MS Windows")
	os.Exit(1)
}
