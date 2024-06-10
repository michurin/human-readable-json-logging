package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/michurin/systemd-env-file/sdenv"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

var (
	debugFlag       = false
	showVersionFlag = false
)

func init() {
	flag.BoolVar(&debugFlag, "d", false, "debug mode")
	flag.BoolVar(&showVersionFlag, "v", false, "show version and exit")
	flag.Parse()
}

func deb(m string) {
	if debugFlag {
		fmt.Println("DEBUG: " + m)
	}
}

const configFile = "pplog.env"

func lookupEnvFile() string {
	cwd, err := os.Getwd()
	if err != nil {
		deb(err.Error())
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		deb(err.Error())
		home = cwd
	}
	for {
		fn := path.Join(cwd, configFile)
		fi, err := os.Stat(fn)
		if err != nil {
			deb(err.Error())
		}
		if err == nil && fi.Mode()&fs.ModeType == 0 {
			deb("file found: " + fn)
			return fn
		}
		cwd = path.Dir(cwd)
		if len(cwd) < len(home) {
			break
		}
	}
	deb("no configuration file has been found")
	return ""
}

func normLine(t string) string {
	return strings.ReplaceAll(strings.TrimSpace(t), "\\e", "\033") + "\n"
}

func showBuildInfo() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("Cannot get build info")
		return
	}
	fmt.Println(info.String())
}

func prepareFormatters() (func([]slogtotext.Pair) error, func([]slogtotext.Pair) error) {
	c := sdenv.NewCollectsion()
	c.PushStd(os.Environ())
	envFile := lookupEnvFile()
	if envFile != "" {
		b, err := os.ReadFile(envFile)
		if err != nil {
			panic(err) // TODO
		}
		pairs, err := sdenv.Parser(b)
		if err != nil {
			panic(err) // TODO
		}
		c.Push(pairs)
	}

	logLine := `{{ .time }} [{{ .level }}] {{ .msg }}{{ range .ALL | rm "time" "level" "msg" }} {{.K}}={{.V}}{{end}}`
	errLine := `INVALID JSON: {{ .TEXT | printf "%q" }}`
	for _, p := range c.Collection() {
		switch p[0] {
		case "PPLOG_LOGLINE":
			logLine = p[1]
		case "PPLOG_ERRLINE":
			errLine = p[1]
		}
	}
	logLine = normLine(logLine)
	errLine = normLine(errLine)
	return slogtotext.MustFormatter(os.Stdout, logLine), slogtotext.MustFormatter(os.Stdout, errLine)
}

func runSubprocessMode() {
	deb("run subprocess mode")

	target := flag.Args()
	binary, err := exec.LookPath(target[0])
	if err != nil {
		panic(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	chFiles := make([]uintptr, 3) // in, out, err
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

func runPipeMode() {
	deb("run pipe mode")
	f, g := prepareFormatters()
	err := slogtotext.Read(os.Stdin, f, g, 32768)
	if err != nil {
		printError(err)
		return
	}
}

func main() {
	if showVersionFlag {
		showBuildInfo()
		return
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		if flag.NArg() >= 1 {
			runSubprocessMode()
		} else {
			runPipeMode()
		}
		close(done)
	}()

	select {
	case <-interrupt:
		deb("interrupt, allow target to shutdown gracefully...")
	case <-done:
	}

	select {
	case <-interrupt:
		deb("second interrupt, exit immediately")
	case <-done:
	}
}

func printError(err error) { // TODO reconsider
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
