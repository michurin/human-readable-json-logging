package slogtotext

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnknownPairs_wanishing(t *testing.T) {
	x := unknowPairs("", map[string]any{
		"b": struct{}{},
		"c": map[string]any{"p": struct{}{}},
		"d": struct{}{},
		"e": map[string]any{"ea": struct{}{}},
	}, map[string]any{
		"a": "A",
		"b": "B", // skip: directly
		"c": "C", // skip: indirect: appears in path
		"d": map[string]any{
			"da": "DA", // doesn't skip: not full path, d only
			"db": "DB", // the same
		},
		"e": map[string]any{
			"ea": "EA", // skip
			"eb": "EB",
		},
	})
	assert.Equal(t, []unknownPair{
		{"a", "A"},
		{"d.da", "DA"},
		{"d.db", "DB"},
		{"e.eb", "EB"},
	}, x)
}

func TestUnknownPairs_types(t *testing.T) {
	data := any(nil)
	d := json.NewDecoder(strings.NewReader(`{"s":"x", "i":1, "t":true, "f":false, "n":null, "a":["x",1], "o":{"x":1}}`))
	d.UseNumber()
	err := d.Decode(&data)
	require.NoError(t, err)
	x := unknowPairs("", nil, data)
	assert.Equal(t, []unknownPair{
		{K: "a.0", V: "x"},
		{K: "a.1", V: "1"},
		{K: "f", V: "false"},
		{K: "i", V: "1"},
		{K: "n", V: "null"},
		{K: "o.x", V: "1"},
		{K: "s", V: "x"},
		{K: "t", V: "true"},
	}, x)
}

func TestUnknownPairs_invalidType(t *testing.T) {
	assert.Equal(t, []unknownPair{{K: "k", V: "UNKNOWN TYPE int8"}}, unknowPairs("k", nil, int8(1)))
}

func TestUnknownPairs_order(t *testing.T) {
	assert.Equal(t, []unknownPair{
		{K: "k.0", V: "a"},
		{K: "k.1", V: "b"},
		{K: "k.2", V: "c"},
		{K: "k.3", V: "d"},
		{K: "k.4", V: "e"},
		{K: "k.5", V: "f"},
		{K: "k.6", V: "g"},
		{K: "k.7", V: "h"},
		{K: "k.8", V: "i"},
		{K: "k.9", V: "j"},
		{K: "k.10", V: "k"},
	}, unknowPairs(
		"", // has to be empty
		nil,
		map[string]any{"k": []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}}),
	)
}
