package slogtotext

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlat(t *testing.T) {
	t.Parallel()
	for i, cs := range []struct {
		in  string
		exp []Pair
	}{{
		in:  `null`, // it skips nulls
		exp: nil,
	}, {
		in:  `true`,
		exp: []Pair{{"NOKEY", "true"}},
	}, {
		in:  `false`,
		exp: []Pair{{"NOKEY", "false"}},
	}, {
		in:  `"x"`,
		exp: []Pair{{"NOKEY", "x"}},
	}, {
		in:  `1`,
		exp: []Pair{{"NOKEY", "1"}},
	}, {
		in:  `[]`,
		exp: []Pair(nil),
	}, {
		in:  `{}`,
		exp: nil,
	}, {
		in: `{"a": null, "b": [1, null, "str", {"p": "q"}], "c": [], "d": true}`,
		exp: []Pair{
			{"b.0", "1"},
			{"b.2", "str"},
			{"b.3.p", "q"},
			{"d", "true"},
		},
	}} {
		cs := cs
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			t.Parallel()
			d := json.NewDecoder(strings.NewReader(cs.in))
			d.UseNumber()
			x := any(nil)
			err := d.Decode(&x)
			require.NoError(t, err)
			r := flat(x)
			assert.Equal(t, cs.exp, r, "in="+cs.in)
		})
	}
	t.Run("invalid_type", func(t *testing.T) {
		require.PanicsWithError(t, "unknown type (pfx=[]) int: 1", func() {
			_ = flat(1)
		})
	})
}
