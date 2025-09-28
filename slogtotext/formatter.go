package slogtotext

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

func tTimeFormatter(args ...any) string {
	// args: fromFormat1, fromFormat2, ... toFormat, timeString
	if len(args) < 3 { //nolint:mnd
		return fmt.Sprintf("Too few arguments: %d", len(args))
	}
	sa := make([]string, len(args))
	for i, v := range args {
		var ok bool
		sa[i], ok = v.(string)
		if !ok {
			return fmt.Sprintf("Invalid type: pos=%[1]d: %[2]T (%[2]v)", i+1, v)
		}
	}
	tm := sa[len(sa)-1] // the last argument is timeString
	tf := sa[len(sa)-2] // the before last argument is target format
	sa = sa[:len(sa)-2] // source formats to try
	for _, v := range sa {
		t, err := time.Parse(v, tm)
		if err != nil {
			e := new(time.ParseError)
			if errors.As(err, &e) {
				continue
			}
			return err.Error() // in fact, it is impossible case, time.Parse always returns *time.ParseError
		}
		return t.Format(tf)
	}
	return tm // return original time as fallback
}

func tRemoveByPfx(args ...any) []Pair { // TODO naive nested loop implementation
	nLast := len(args) - 1
	if nLast <= 0 {
		panic(fmt.Sprintf("Invalid number of args: %d: %v", len(args), args))
	}
	c := make([]string, nLast)
	ok := false
	for i := 0; i < nLast; i++ {
		c[i], ok = args[i].(string)
		if !ok {
			panic(fmt.Sprintf("Invalid type: idx=%[1]d: %[2]T: %[2]v", i, args[i]))
		}
	}
	av := args[nLast]
	a, ok := av.([]Pair)
	if !ok {
		panic(fmt.Sprintf("Invalid type: %[1]T: %[1]v: only .ALL allows", av))
	}
	r := []Pair(nil)
	for _, x := range a {
		found := false
		for _, p := range c {
			if strings.HasPrefix(x.K, p) {
				found = true
				break
			}
		}
		if !found {
			r = append(r, x)
		}
	}
	return r
}

func tRemove(args ...any) []Pair {
	nLast := len(args) - 1
	if nLast <= 0 {
		panic(fmt.Sprintf("Invalid number of args: %d: %v", len(args), args))
	}
	c := make(map[string]struct{}, nLast)
	for i := 0; i < nLast; i++ {
		s, ok := args[i].(string)
		if !ok {
			panic(fmt.Sprintf("Invalid type: idx=%[1]d: %[2]T: %[2]v", i, args[i]))
		}
		c[s] = struct{}{}
	}
	av := args[nLast]
	a, ok := av.([]Pair)
	if !ok {
		panic(fmt.Sprintf("Invalid type: %[1]T: %[1]v: only .ALL allows", av))
	}
	r := []Pair(nil)
	for _, x := range a {
		if _, ok := c[x.K]; !ok {
			r = append(r, x)
		}
	}
	return r
}

func tTrimSpace(args ...any) string {
	r := make([]string, len(args))
	for i, v := range args {
		if s, ok := v.(string); ok {
			r[i] = strings.TrimSpace(s)
		} else {
			r[i] = fmt.Sprintf("%#v", v)
		}
	}
	return strings.Join(r, " ")
}

func tSkipLineIf(x *atomic.Bool, xor bool) func(args ...any) string {
	return func(args ...any) string {
		f := false
		for _, v := range args {
			switch x := v.(type) {
			case bool:
				f = f || x
			case string:
				f = f || (len(x) > 0)
			case int:
				f = f || (x != 0)
			}
			if f {
				break
			}
		}
		x.Store(f != xor) // xor operation
		return ""
	}
}

func Formatter(stream io.Writer, templateString string) (func([]Pair) error, error) {
	flag := new(atomic.Bool)
	tm, err := template.
		New("base").
		Option("missingkey=zero").
		Funcs(template.FuncMap{
			"tmf":            tTimeFormatter,
			"rm":             tRemove,
			"rmByPfx":        tRemoveByPfx,
			"xjson":          tXJson,
			"xxjson":         tXXJson,
			"trimSpace":      tTrimSpace,
			"skipLineIf":     tSkipLineIf(flag, false),
			"skipLineUnless": tSkipLineIf(flag, true),
		}).
		Funcs(sprig.FuncMap()).
		Parse(templateString)
	if err != nil {
		return nil, err // TODO wrap?
	}

	return func(p []Pair) error {
		kv := make(map[string]any, len(p))
		for _, v := range p {
			kv[v.K] = v.V
		}
		q := make([]Pair, 0, len(p))
		for _, v := range p {
			if v.K != rawInputKey {
				q = append(q, v)
			}
		}
		sort.Slice(q, func(i, j int) bool { return q[i].K < q[j].K })
		kv[allKey] = q
		buff := new(bytes.Buffer)
		err := tm.Execute(buff, kv)
		if err != nil {
			return err // TODO wrap error?
		}
		if !flag.Load() {
			_, err = io.Copy(stream, buff)
			if err != nil {
				return err // TODO wrap error?
			}
		}
		return nil
	}, nil
}

func MustFormatter(stream io.Writer, templateString string) func([]Pair) error {
	f, err := Formatter(stream, templateString)
	if err != nil {
		panic(err.Error()) // TODO wrap?
	}
	return f
}
