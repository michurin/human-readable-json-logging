package slogtotext

import (
	"bytes"
	"encoding/json"
)

func fixAny(x any) any {
	switch q := x.(type) {
	case map[string]any:
		r := make(map[string]any, len(q))
		for k, v := range q {
			r[k] = fixAny(v)
		}
		return r
	case []any:
		r := make([]any, len(q))
		for i, e := range q {
			r[i] = fixAny(e)
		}
		return r
	case string:
		if a, ok := strToAny(q); ok {
			return a
		}
	}
	return x
}

func strToAny(x string) (any, bool) {
	d := []byte(x)
	if !json.Valid(d) {
		return nil, false
	}
	q := any(nil)
	buf := bytes.NewBuffer(d)
	dec := json.NewDecoder(buf)
	dec.UseNumber()
	err := dec.Decode(&q)
	if err != nil {
		return nil, false
	}
	return fixAny(q), true
}

func tXJson(x string) string {
	p, ok := strToAny(x)
	if !ok {
		return x
	}
	b, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
