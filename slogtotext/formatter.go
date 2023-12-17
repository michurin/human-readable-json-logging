package slogtotext

import (
	"fmt"
	"io"
	"text/template"
	"text/template/parse"
)

func tRemove(args ...any) []Pair {
	if len(args) < 1 {
		panic(fmt.Sprintf("Invalid number of args: %d: %v", len(args), args))
	}
	c := map[string]struct{}{}
	for i := 0; i < len(args)-1; i++ {
		s, ok := args[i].(string)
		if !ok {
			panic(fmt.Sprintf("Invalid type: %[1]T: %[1]v", s))
		}
		c[s] = struct{}{}
	}
	av := args[len(args)-1]
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

func Formatter(stream io.Writer, templateString string) func([]Pair) error {
	if len(templateString) > 0 && templateString[len(templateString)-1] != '\n' {
		templateString += "\n"
	}

	tm, err := template.New("x").Option("missingkey=zero").Funcs(template.FuncMap{"remove": tRemove}).Parse(templateString)
	if err != nil {
		return func([]Pair) error {
			return err // TODO wrap?
		}
	}

	knownKeys := map[string]struct{}{}
	for _, x := range tm.Root.Nodes {
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

	return func(p []Pair) error {
		kv := make(map[string]any, len(p))
		u := []Pair(nil)
		for _, v := range p {
			kv[v.K] = v.V
			if _, ok := knownKeys[v.K]; !ok {
				u = append(u, v)
			}
		}
		kv["ALL"] = p
		kv["UNKNOWN"] = u
		err := tm.Execute(stream, kv)
		if err != nil {
			panic(err) // TODO
			return err // TODO wrap error?
		}
		return nil
	}
}
