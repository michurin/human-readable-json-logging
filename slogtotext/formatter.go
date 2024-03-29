package slogtotext

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"
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
	if nLast < 0 {
		panic(fmt.Sprintf("Invalid number of args: %d: %v", len(args), args))
	}
	c := make([]string, nLast)
	ok := false
	for i := 0; i < nLast; i++ {
		c[i], ok = args[i].(string)
		if !ok {
			panic(fmt.Sprintf("Invalid type: idx=%d: %[1]T: %[1]v", i, args[i]))
		}
	}
	av := args[nLast]
	a, ok := av.([]Pair)
	if !ok {
		panic(fmt.Sprintf("Invalid type: %[1]T: %[1]v", av))
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
	if nLast < 0 {
		panic(fmt.Sprintf("Invalid number of args: %d: %v", len(args), args))
	}
	c := make(map[string]struct{}, nLast)
	for i := 0; i < nLast; i++ {
		s, ok := args[i].(string)
		if !ok {
			panic(fmt.Sprintf("Invalid type: %[1]T: %[1]v", s))
		}
		c[s] = struct{}{}
	}
	av := args[nLast]
	a, ok := av.([]Pair)
	if !ok {
		panic(fmt.Sprintf("Invalid type: %[1]T: %[1]v", av))
	}
	r := []Pair(nil)
	for _, x := range a {
		if _, ok := c[x.K]; !ok {
			r = append(r, x)
		}
	}
	return r
}

func Formatter(stream io.Writer, templateString string) (func([]Pair) error, error) {
	tm, err := template.New("x").Option("missingkey=zero").Funcs(template.FuncMap{
		"tmf":     tTimeFormatter,
		"rm":      tRemove,
		"rmByPfx": tRemoveByPfx,
		"xjson":   tXJson,
		"xxjson":  tXXJson,
	}).Parse(templateString)
	if err != nil {
		return nil, err // TODO wrap?
	}

	return func(p []Pair) error {
		kv := make(map[string]any, len(p))
		for _, v := range p {
			kv[v.K] = v.V
		}
		kv["ALL"] = p
		err := tm.Execute(stream, kv)
		if err != nil {
			return err // TODO wrap error?
		}
		return nil
	}, nil
}

func FormatterMust(stream io.Writer, templateString string) func([]Pair) error {
	f, err := Formatter(stream, templateString)
	if err != nil {
		panic(err.Error()) // TODO wrap?
	}
	return f
}
