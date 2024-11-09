package slogtotext

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

func tTimeFormatter(from, to string, tm any) string {
	ts, ok := tm.(string)
	if !ok {
		return fmt.Sprintf("invalid time type: %[1]T (%[1]v)", tm)
	}
	t, err := time.Parse(from, ts)
	if err != nil {
		return err.Error()
	}
	return t.Format(to)
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

func Formatter(stream io.Writer, templateString string) (func([]Pair) error, error) {
	tm, err := template.
		New("base").
		Option("missingkey=zero").
		Funcs(template.FuncMap{
			"tmf":       tTimeFormatter,
			"rm":        tRemove,
			"rmByPfx":   tRemoveByPfx,
			"xjson":     tXJson,
			"xxjson":    tXXJson,
			"trimSpace": tTrimSpace,
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
		err := tm.Execute(stream, kv)
		if err != nil {
			return err // TODO wrap error?
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
