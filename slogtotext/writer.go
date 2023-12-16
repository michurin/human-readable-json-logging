// Package slogtotext provides wrapper for io.Writer interface to reshape JSON structured logs according text/template-templates.
package slogtotext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"text/template"
	"text/template/parse"
	"time"
)

type pplog struct {
	mx             *sync.Mutex
	next           io.Writer
	collecting     bool
	buff           []byte
	logline        *template.Template
	errline        *template.Template
	knownKeys      map[string]any
	maxParsibleLen int
}

func (pp *pplog) Write(p []byte) (int, error) {
	pp.mx.Lock()
	defer pp.mx.Unlock()
	n := len(p)
	for len(p) > 0 {
		s := bytes.IndexByte(p, '\n')
		if s == -1 {
			err := pp.acc(p)
			if err != nil {
				return 0, err // hmm... zero?
			}
			break
		} else {
			err := pp.fin(p[:s+1])
			if err != nil {
				return 0, err
			}
			p = p[s+1:]
		}
	}
	return n, nil
}

func (pp *pplog) fin(p []byte) error {
	err := pp.acc(p)
	if err != nil {
		return err
	}
	if pp.collecting {
		if !json.Valid(pp.buff) { // Decode stops on success result not on end of data, so we have to check is whole buffer is valid, see json.Unmashal
			err = pp.errline.Execute(pp.next, string(pp.buff[:len(pp.buff)-1])) // here in fin() we are sure that we have \n and the end of the buffer [fragile]
			pp.buff = pp.buff[:0]
			if err != nil {
				return fmt.Errorf("errline template: %w", err)
			}
			return nil
		}
		// https://github.com/golang/go/issues/24963
		data := any(nil)
		d := json.NewDecoder(bytes.NewReader(pp.buff))
		d.UseNumber()
		err := d.Decode(&data)
		if err != nil {
			return fmt.Errorf("decode: %w", err) // impossible
		}
		pp.buff = pp.buff[:0]
		mdata, ok := data.(map[string]any)
		if ok {
			mdata["UNKNOWN"] = unknowPairs("", pp.knownKeys, data)
			data = mdata
		}
		err = pp.logline.Execute(pp.next, data)
		if err != nil {
			return fmt.Errorf("logline template: %w", err)
		}
	}
	pp.collecting = true
	return nil
}

func (pp *pplog) acc(p []byte) error {
	if len(pp.buff)+len(p) > pp.maxParsibleLen {
		pp.collecting = false
		err := pp.flush()
		if err != nil {
			return err
		}
	}
	if pp.collecting {
		pp.buff = append(pp.buff, p...)
		return nil
	}
	_, err := pp.next.Write(p)
	if err != nil {
		return err
	}
	return nil
}

func (pp *pplog) flush() error { // TODO it is used in acc only
	if len(pp.buff) == 0 {
		return nil
	}
	_, err := pp.next.Write(pp.buff)
	pp.buff = pp.buff[:0]
	if err != nil {
		return err
	}
	return nil
}

func PPLog(
	writer io.Writer,
	errlineTemplate,
	loglineTemplate string,
	knownKeys map[string]any,
	funcMap map[string]any,
	maxParsibleLen int,
) io.Writer {
	// TODO: validate knownKeys
	fm := template.FuncMap{"tmf": func(from, to string, tm any) string {
		ts, ok := tm.(string)
		if !ok {
			return fmt.Sprintf("invalid time type: %[1]T (%[1]v)", tm)
		}
		t, err := time.Parse(from, ts)
		if err != nil {
			return err.Error()
		}
		return t.Format(to)
	}}
	for k, v := range funcMap {
		fm[k] = v
	}
	if len(errlineTemplate) == 0 {
		errlineTemplate = `INVALID JSON: {{. | printf "%q"}}`
	}
	if len(loglineTemplate) == 0 {
		loglineTemplate = `{{.time}} [{{.level}}] {{.msg}}{{range .UNKNOWN}} {{.K}}={{.V}}{{end}}`
	}
	ll := template.Must(template.New("l").Option("missingkey=zero").Funcs(fm).Parse(loglineTemplate + "\n"))
	el := template.Must(template.New("e").Option("missingkey=zero").Funcs(fm).Parse(errlineTemplate + "\n"))
	if knownKeys == nil {
		knownKeys = map[string]any{}
		for _, x := range ll.Root.Nodes {
			if n, ok := x.(*parse.ActionNode); ok {
				for _, c := range n.Pipe.Cmds {
					for _, a := range c.Args {
						if b, ok := a.(*parse.FieldNode); ok {
							if len(b.Ident) > 0 {
								knownKeys[b.Ident[0]] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	if maxParsibleLen <= 0 {
		maxParsibleLen = 0x4000 // 16k
	}
	return &pplog{
		mx:             new(sync.Mutex),
		next:           writer,
		collecting:     true,
		buff:           make([]byte, 0, maxParsibleLen),
		logline:        ll,
		errline:        el,
		knownKeys:      knownKeys,
		maxParsibleLen: maxParsibleLen,
	}
}
