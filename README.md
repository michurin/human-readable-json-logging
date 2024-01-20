# human-readable-json-logging

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

```sh
pplog service
# or even
pplog go run ./cmd/service/...
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

```
{"type":"I","time":"2024-01-01T07:33:44Z","message":"RPC call","k8s_node":"ix-x-kub114","k8s_pod":"booking-v64-64cf64db6d-gm9pc","cluster_name":"zeta","env":"prod","tag":"service.booking","lineno":39,"function":"xxx.xx/service-booking/internal/rpc/booking.(*Handler).Handle.func1","run":"578710a04dbb","comp":"rpc.booking","payload_resp":"{\"provider\":\"None\"}","payload_req":"{\"userId\":34664834}","xsource":"profile","_tracing":{"uber-trace-id":"669f:6a2c:c35c:1"}}
```

It turns to:

```
I 07:33:44 5787 rpc.booking xxx.xx/service-booking/internal/rpc/booking.(*Handler).Handle.func1 39 RPC call payload_req={"userId":34664834} payload_resp={"provider":"None"}
```

## TODO

- Add original line to set of template substitutions even if it has been parsed successfully
- Godoc

## Known issues

### Not optimal integration with log/slog

If you decided to use this code as library as part of your product, you have to keep in mind, that
this tool provides `io.Writer` to pipe log stream. It is easiest way to modify behavior of logger, however
it leads to overhead for extra marshaling/unmarshaling. However, as well as we use human readable logs in
local environment only, it is acceptable to have a little overhead.

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
