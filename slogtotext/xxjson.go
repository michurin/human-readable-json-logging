package slogtotext

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
)

type kv struct {
	k string
	v any
}

type kvSlice []kv

func (x kvSlice) Len() int           { return len(x) }
func (x kvSlice) Less(i, j int) bool { return x[i].k < x[j].k }
func (x kvSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type writer struct {
	w io.Writer
}

func (w *writer) wr(x ...[]byte) {
	for _, e := range x {
		_, err := w.w.Write(e)
		if err != nil {
			panic(err.Error())
		}
	}
}

func marshal(x any, buf *writer, clr *xxJSONColors) { // ACHTUNG we do not care about looping
	switch q := x.(type) {
	case map[string]any:
		kk := make(kvSlice, 0, len(q))
		for k, v := range q {
			kk = append(kk, kv{k: k, v: v})
		}
		sort.Sort(kk)
		buf.wr([]byte(`{`))
		for i, e := range kk {
			if i > 0 {
				buf.wr([]byte(`,`))
			}
			buf.wr(clr.KeyOpen, []byte(e.k), clr.KeyClose, []byte(`:`))
			marshal(e.v, buf, clr)
		}
		buf.wr([]byte(`}`))
	case []any:
		buf.wr([]byte(`[`))
		for i, e := range q {
			if i > 0 {
				buf.wr([]byte(`,`))
			}
			marshal(e, buf, clr)
		}
		buf.wr([]byte(`]`))
	case string:
		buf.wr(clr.StringOpen, []byte(q), clr.StringClose)
	case bool:
		if q {
			buf.wr(clr.TrueOpen, []byte(`T`), clr.TrueClose)
		} else {
			buf.wr(clr.FalseOpen, []byte(`F`), clr.FalseClose)
		}
	case nil:
		buf.wr(clr.NullOpen, []byte(`N`), clr.NullClose)
	case json.Number:
		buf.wr(clr.NumberOpen, []byte(q.String()), clr.NumberClose)
	}
}

type xxJSONColors struct {
	KeyOpen     []byte
	KeyClose    []byte
	FalseOpen   []byte
	FalseClose  []byte
	TrueOpen    []byte
	TrueClose   []byte
	NullOpen    []byte
	NullClose   []byte
	StringOpen  []byte
	StringClose []byte
	NumberOpen  []byte
	NumberClose []byte
}

var defaultColors = xxJSONColors{ //nolint:gochecknoglobals
	KeyOpen:     []byte("\033[93m"),
	KeyClose:    []byte("\033[0m"),
	FalseOpen:   []byte("\033[91m"),
	FalseClose:  []byte("\033[0m"),
	TrueOpen:    []byte("\033[92m"),
	TrueClose:   []byte("\033[0m"),
	NullOpen:    []byte("\033[95m"),
	NullClose:   []byte("\033[0m"),
	StringOpen:  []byte("\033[35m"),
	StringClose: []byte("\033[0m"),
	NumberOpen:  []byte("\033[95m"),
	NumberClose: []byte("\033[0m"),
}

func tXXJson(x string) string {
	p, ok := strToAny(x)
	if !ok {
		return x
	}
	buf := bytes.NewBuffer(nil)
	marshal(p, &writer{w: buf}, &defaultColors) // TODO configurable colors
	return buf.String()
}
