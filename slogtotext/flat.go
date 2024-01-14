package slogtotext

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Pair struct {
	K string
	V string
}

type pair struct {
	k string
	v any
}

func fill(m *[]Pair, pfx []string, d any) {
	switch x := d.(type) {
	case nil:
	case bool:
		t := "false"
		if x {
			t = "true"
		}
		*m = append(*m, Pair{K: kjoin(pfx), V: t})
	case string:
		*m = append(*m, Pair{K: kjoin(pfx), V: x})
	case json.Number:
		*m = append(*m, Pair{K: kjoin(pfx), V: x.String()})
	case []any:
		for i, v := range x {
			fill(m, append(pfx, strconv.Itoa(i)), v)
		}
	case map[string]any:
		kv := make([]pair, len(x))
		n := 0
		for k, v := range x {
			kv[n].k = k
			kv[n].v = v
			n++
		}
		sort.Slice(kv, func(i, j int) bool { return kv[i].k < kv[j].k })
		for _, e := range kv {
			fill(m, append(pfx, e.k), e.v)
		}
	default:
		panic(fmt.Errorf("unknown type (pfx=%[2]s) %[1]T: %[1]v", x, pfx))
	}
}

func kjoin(pfx []string) string {
	if len(pfx) == 0 {
		return "NOKEY" // in case the root element is not object or array
	}
	return strings.Join(pfx, ".")
}

func flat(d any) []Pair {
	res := []Pair(nil)
	fill(&res, nil, d)
	return res
}
