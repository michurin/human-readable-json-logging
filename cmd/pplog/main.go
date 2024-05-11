package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/michurin/systemd-env-file/sdenv"
	"golang.org/x/sync/errgroup"

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

func main() {
	if showVersionFlag {
		showBuildInfo()
		return
	}
	if flag.NArg() < 2 {
		fmt.Println("Usage: pplog [-d] [-v] your_command arg arg arg...")
		return
	}
	args := flag.Args()[1:]
	command := flag.Args()[0]

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

	ctx := context.Background()

	errGrp, ctx := errgroup.WithContext(ctx)

	rd, wr := io.Pipe()

	f := slogtotext.MustFormatter(os.Stdout, logLine)
	g := slogtotext.MustFormatter(os.Stdout, errLine)
	errGrp.Go(func() error {
		return slogtotext.Read(rd, f, g, 32768)
	})

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = wr
	cmd.Stderr = wr
	deb("running: " + cmd.String())
	errGrp.Go(func() error {
		err := cmd.Run()
		rd.Close()
		return err
	})

	err := errGrp.Wait()
	deb("running fin")
	if err != nil {
		printError(err)
	}
	exitCode := cmd.ProcessState.ExitCode() // exit code can be set even if error
	if exitCode < 0 {
		exitCode = 1
	}
	os.Exit(exitCode)
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
