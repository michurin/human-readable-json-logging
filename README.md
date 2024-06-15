# human-readable-json-logging (pplog)

[![build](https://github.com/michurin/human-readable-json-logging/actions/workflows/ci.yaml/badge.svg)](https://github.com/michurin/human-readable-json-logging/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/michurin/human-readable-json-logging/graph/badge.svg?token=LDYMK3ZM06)](https://codecov.io/gh/michurin/human-readable-json-logging)
[![Go Report Card](https://goreportcard.com/badge/github.com/michurin/human-readable-json-logging)](https://goreportcard.com/report/github.com/michurin/human-readable-json-logging)

It is library and binary CLI tool to make JSON logs readable including colorizing.
The formatting is based on standard goland templating engine [text/template](https://pkg.go.dev/text/template).

The CLI tool obtains templates from environment or configuration file. See examples below.

You can find examples of using the library in [documentation](https://pkg.go.dev/github.com/michurin/human-readable-json-logging/slogtotext).
Long story short, you are to direct output of your JSON logger (including modern [slog](https://pkg.go.dev/log/slog)) to magic reader and
readable loglines shows up.

## Install and use

```sh
go install -v github.com/michurin/human-readable-json-logging/cmd/...@latest
```

Running in subprocess mode:

```sh
pplog ./service
# or even
pplog go run ./cmd/service/...
```

Running in pipe mode:

```sh
./service | pplog
# or with redirections if you need to take both stderr and stdout
./service 2>&1 | pplog
# or the same redirections in modern shells
./service |& pplog
```

## Real life example

One of my configuration file:

```sh
# File pplog.env. The syntax of file is right the same as systemd env-files.
# You can put it into your working dirrectory or any parrent.
# You are free to set this variables in your .bashrc as well.

PPLOG_LOGLINE='
{{- if .type     }}{{ if eq .type "I" }}\e[92m{{ end }}{{if eq .type "E" }}\e[1;33;41m{{ end }}{{.type}}\e[0m {{ end }}
{{- if .time     }}{{ if eq .type "E" }}\e[93;41m{{ else }}\e[33m{{ end }}{{.time | tmf "2006-01-02T15:04:05Z07:00" "15:04:05"}}\e[0m {{ end }}
{{- if .run      }}\e[93m{{ .run | printf "%4.4s"}}\e[0m {{ end }}
{{- if .comp     }}\e[92m{{ .comp     }}\e[0m {{ end }}
{{- if .scope    }}\e[32m{{ .scope    }}\e[0m {{ end }}
{{- if .ci_test_name }}\e[35;44;1m{{ .ci_test_name}}\e[0m {{ end }}
{{- if .function }}\e[94m{{ .function }} \e[95m{{.lineno}}\e[0m {{ end }}
{{- if .message  }}\e[97m{{ .message  }}\e[0m {{ end }}
{{- if .error    }}\e[91m{{ .error    }}\e[0m {{ end }}
{{- if .error_trace }}\e[93m{{ .error_trace }}\e[0m {{ end }}
{{- range .ALL | rmByPfx
    "_tracing"
    "ci_test_name"
    "cluster_name"
    "comp"
    "env"
    "error"
    "error_trace"
    "function"
    "k8s_"
    "lineno"
    "message"
    "run"
    "scope"
    "tag"
    "time"
    "type"
    "xsource"
}}\e[33m{{.K}}\e[0m={{.V}} {{ end }}
'

PPLOG_ERRLINE='{{ if .binary }}{{ .text }}{{ else }}\e[97m{{ .text }}\e[0m{{ end }}'
```

My original logs look like this:

```json
{"type":"I","time":"2024-01-01T07:33:44Z","message":"RPC call","k8s_node":"ix-x-kub114","k8s_pod":"booking-v64-64cf64db6d-gm9pc","cluster_name":"zeta","env":"prod","tag":"service.booking","lineno":39,"function":"xxx.xx/service-booking/internal/rpc/booking.(*Handler).Handle.func1","run":"578710a04dbb","comp":"rpc.booking","payload_resp":"{\"provider\":\"None\"}","payload_req":"{\"userId\":34664834}","xsource":"profile","_tracing":{"uber-trace-id":"669f:6a2c:c35c:1"}}
```

It turns to:

```
I 07:33:44 5787 rpc.booking xxx.xx/service-booking/internal/rpc/booking.(*Handler).Handle.func1 39 RPC call payload_req={"userId":34664834} payload_resp={"provider":"None"}
```

## Step by step customization

First things first, I recommend you to prepare small file with your logs. Let's name it `example.log`.

Now you can start to play with it by command like that:

```sh
pplog cat example.log
```

You will see some formatted logs.

Now create your first `pplog.env`. You can start from this universal one:

```sh
PPLOG_LOGLINE='{{range .ALL}}{{.K}}={{.V}} {{end}}'
PPLOG_ERRLINE='Invalid JONS: {{.text}}'
```

You will see all your logs in KEY=VALUE format. Now look over all your keys and choose one you want
to see in the first place. Say, `message`. Modify your `pplog.env` this way:

```sh
PPLOG_LOGLINE='{{.message}}{{range .ALL | rm "message"}} {{.K}}={{.V}}{{end}}'
PPLOG_ERRLINE='Invalid JONS: {{.text}}'
```

You will see `message` in the first place and remove it from KEY=VALUE tail.

Now, you are free to add colors:

```sh
PPLOG_LOGLINE='\e[32m{{.message}}\e[m{{range .ALL | rm "message"}} {{.K}}={{.V}}{{end}}'
PPLOG_ERRLINE='Invalid JONS: {{.text}}'
```

We makes `message` green. Keep shaping your logs field by field.

## Template functions

- `tmf`
- `rm`
- `rmByPfx`
- `xjson`
- `xxjson` (experimental)

## Template special variables

- In `PPLOG_LOGLINE` template:
    - `RAW_INPUT`
    - `ALL` — list of all pairs key-value
- In `PPLOG_ERRLINE` template:
    - `TEXT`
    - `BINARY` — does TEXT contains control characters
- If `PPLOG_CHILD_MODE` not empty `pplog` runs in child mode as if it has `-c` switch

## Most common colors

```
Text colors          Text High            Background           Hi Background         Decoration
------------------   ------------------   ------------------   -------------------   --------------------
\e[30mBlack  \e[0m   \e[90mBlack  \e[0m   \e[40mBlack  \e[0m   \e[100mBlack  \e[0m   \e[1mBold      \e[0m
\e[31mRed    \e[0m   \e[91mRed    \e[0m   \e[41mRed    \e[0m   \e[101mRed    \e[0m   \e[4mUnderline \e[0m
\e[32mGreen  \e[0m   \e[92mGreen  \e[0m   \e[42mGreen  \e[0m   \e[102mGreen  \e[0m   \e[7mReverse   \e[0m
\e[33mYellow \e[0m   \e[93mYellow \e[0m   \e[43mYellow \e[0m   \e[103mYellow \e[0m
\e[34mBlue   \e[0m   \e[94mBlue   \e[0m   \e[44mBlue   \e[0m   \e[104mBlue   \e[0m   Combinations
\e[35mMagenta\e[0m   \e[95mMagenta\e[0m   \e[45mMagenta\e[0m   \e[105mMagenta\e[0m   -----------------------
\e[36mCyan   \e[0m   \e[96mCyan   \e[0m   \e[46mCyan   \e[0m   \e[106mCyan   \e[0m   \e[1;4;103;31mWARN\e[0m
\e[37mWhite  \e[0m   \e[97mWhite  \e[0m   \e[47mWhite  \e[0m   \e[107mWhite  \e[0m
```

## TODO

- Usage: show templates in debug mode
- Behavior tests:
    - `-c`
    - `PPLOG_CHILD_MODE` environment variable
    - basic `runs-on: windows-latest`
    - passing exit code
- Docs: explain main features of binary: modes etc.
- Docs: link to console control codes info
- Docs: write template functions guide and examples
- Docs: contributing guide: how to run behavior tests locally
- Docs: godoc

## Known issues

### Not optimal integration with log/slog

If you decided to use this code as library as part of your product, you have to keep in mind, that
this tool provides `io.Writer` to pipe log stream. It is easiest way to modify behavior of logger, however
it leads to overhead for extra marshaling/unmarshaling. However, as well as we use human readable logs in
local environment only, it is acceptable to have a little overhead.

### Subprocesses handling issues

The problem is that many processes have to be synchronized: shell-process, pplog-process, target-process with all its children...

Long story short, the most reliable mode is pipe mode:

```sh
./service |& pplog
```

TODO: Explain subprocess-mode and child-mode (`-c`)

### Line-by-line processing

This tool processes input stream line by line. It means, that it won't work with multiline JSONs in logs like that

```json
{
  "level": "info",
  "message": "obtaining data"
}
{
  "level": "error",
  "message": "invalid data"
}
```

as well as it won't work with mixed lines like this:

```
Raw message{"message": "valid json log record"}
```

all that cases will be considered and output as wrong JSON.

Honestly, I have tried to implement smarter scanner. It's not a big deal, however,
in fact, it is not convenient. For instance, it consider message like this `code=200`
in wearied way: `code=` is wrong JSON, however `200` is valid JSON.
Things like that `0xc00016f0e1` get really awful.

I have played with different approaches
and decided just to split logs line by line first.

## Contributors

- [vitalyshatskikh](https://github.com/vitalyshatskikh)
