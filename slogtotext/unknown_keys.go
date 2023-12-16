package slogtotext

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type unknownPair struct {
	K string
	V string
}

func unknowPairs(prefix string, knownKeys any, data any) []unknownPair {
	res := []unknownPair(nil)
	switch data := data.(type) {
	case map[string]any:
		if len(prefix) > 0 {
			prefix += "."
		}
		for _, k := range orderedKeys(data) {
			kk := any(nil)
			if p, ok := knownKeys.(map[string]any); ok {
				kk = p[k]
			}
			res = append(res, unknowPairs(prefix+k, kk, data[k])...)
		}
	case []any:
		if len(prefix) > 0 {
			prefix += "."
		}
		for i, e := range data {
			res = append(res, unknowPairs(prefix+strconv.Itoa(i), knownKeys, e)...) // hmm... knownKeys[something]?
		}
	case string:
		if knownKeys == nil {
			res = append(res, unknownPair{K: prefix, V: data})
		}
	case json.Number:
		if knownKeys == nil {
			res = append(res, unknownPair{K: prefix, V: data.String()})
		}
	case bool:
		if knownKeys == nil {
			res = append(res, unknownPair{K: prefix, V: boolString(data)})
		}
	case nil:
		if knownKeys == nil {
			res = append(res, unknownPair{K: prefix, V: "null"})
		}
	default:
		res = append(res, unknownPair{K: prefix, V: fmt.Sprintf("UNKNOWN TYPE %T", data)}) // impossible
	}
	return res
}

func boolString(x bool) string {
	if x {
		return "true"
	}
	return "false"
}

func orderedKeys(data map[string]any) []string {
	kk := make([]string, 0, len(data))
	for k := range data {
		kk = append(kk, k)
	}
	sort.Strings(kk)
	return kk
}
