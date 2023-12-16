package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`Usage: pplog your_command arg arg arg...`)
		return
	}
	args := os.Args[1:]
	pplog := slogtotext.PPLog(
		os.Stdout,
		`{{. | invalid}}`,
		`
{{- if eq .type "I" }}`+"\033[92m"+`{{end -}}
{{- if eq .type "E" }}`+"\033[1;33;41m"+`{{end -}}
{{- .type -}}
{{- if eq .type "I" }}`+"\033[0m"+`{{end}} {{.time | tmf "2006-01-02T15:04:05Z07:00" "15:04:05" }}`+"\033[0m"+
			`{{if .run}} `+"\033[93m"+`{{.run | printf "%4.4s"}}`+"\033[0m"+`{{end}}`+
			`{{if .comp}} `+"\033[92m"+`{{.comp}}`+"\033[0m"+`{{end}}`+
			`{{if .scope}} `+"\033[32m"+`{{.scope}}`+"\033[0m"+`{{end}}`+
			`{{if .ci_test_name}} `+"\033[35;44;1m"+`{{.ci_test_name}}`+"\033[0m"+`{{end}}`+
			" \033[94m"+`{{.function}} {{.lineno}}`+"\033[39m"+
			" \033[97m{{.message}}\033[0m"+
			`{{range .UNKNOWN}} `+"\033[93m"+`{{.K}}`+"\033[39m"+`={{.V}}{{end}}`+"\033[0m",
		map[string]any{
			"_tracing":     map[string]any{"uber-trace-id": struct{}{}},
			"ci_test_name": struct{}{},
			"cluster_name": struct{}{},
			"comp":         struct{}{},
			"env":          struct{}{},
			"function":     struct{}{},
			"lineno":       struct{}{},
			"message":      struct{}{},
			"run":          struct{}{},
			"scope":        struct{}{},
			"tag":          struct{}{},
			"time":         struct{}{},
			"type":         struct{}{},
			"xsource":      struct{}{},
		},
		map[string]any{
			"invalid": func(x string) string {
				i := strings.Index(x, "FAIL")
				if i >= 0 {
					x = x[:i] + "\033[41;93;1mFAIL\033[0m\033[97m" + x[i+4:]
				}
				return "\033[97m" + x + "\033[0m"
			},
		},
		16384,
	)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = pplog
	cmd.Stderr = pplog
	// signal.Ignore(syscall.SIGPIPE) // is it really needed?
	err := cmd.Run()
	if err != nil {
		printError(err)
	}
	exitCode := cmd.ProcessState.ExitCode() // exit code can be set even if error
	if exitCode < 0 {
		exitCode = 1
	}
	os.Exit(exitCode)
}

func printError(err error) {
	pe := new(os.PathError)
	if errors.As(err, &pe) {
		if pe.Err == syscall.EBADF { // fragile code; somehow syscall.Errno.Is doesn't recognize EBADF, so we unable to use errors.As
			// maybe it is good idea just ignore SIGPIPE
			fmt.Fprintf(os.Stderr, "PPLog: It seems output descriptor has been closed\n") // trying to report it to stderr
			return
		}
	}
	xe := new(exec.ExitError)
	if errors.As(err, &xe) {
		fmt.Printf("exit code = %d: %s\n", xe.ExitCode(), xe.Error()) // just for information
		return
	}
	fmt.Printf("Error: %s\n", err.Error())
}
