package slogtotext_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/michurin/human-readable-json-logging/slogtotext"
)

type splitReader struct {
	size int
	data []byte
}

func (sr *splitReader) Read(b []byte) (int, error) {
	if len(sr.data) == 0 {
		return 0, io.EOF
	}
	c := len(b)
	if sr.size < c {
		c = sr.size
	}
	if len(sr.data) < c {
		c = len(sr.data)
	}
	copy(b, sr.data[:c])
	sr.data = sr.data[c:]
	return c, nil
}

func pairs(p []slogtotext.Pair) string {
	x := make([]string, len(p))
	for i, v := range p {
		x[i] = v.K + "=" + v.V
	}
	return strings.Join(x, "\x20")
}

func collector(t *testing.T) (
	func([]slogtotext.Pair) error,
	func([]slogtotext.Pair) error,
	func() string,
) {
	out := []string(nil)
	f := func(p []slogtotext.Pair) error {
		out = append(out, pairs(p))
		return nil
	}
	g := func(p []slogtotext.Pair) error {
		t.Helper()
		require.Len(t, p, 1)
		require.Equal(t, "text", p[0].K)
		out = append(out, fmt.Sprintf("%q", p[0].V))
		return nil
	}
	c := func() string {
		return strings.Join(out, "|")
	}
	return f, g, c
}

func TestAll(t *testing.T) {
	t.Parallel()
	data := `{"pid": 12} }invalid{ {"next": true}`
	{
		data := data
		t.Run("justCheckOnNativeStringsReader", func(t *testing.T) {
			t.Parallel()
			buf := strings.NewReader(data)
			f, g, c := collector(t)
			err := slogtotext.Read(buf, f, g, 1000)
			require.NoError(t, err)
			assert.Equal(t, `pid=12|" }invalid{"|next=true`, c())
		})
	}

	expectations := map[int]string{
		1:  `"{"|"\""|"p"|"i"|"d"|"\""|":"|" "|=1|=2|"}"|" "|"}"|"i"|"n"|"v"|"a"|"l"|"i"|"d"|"{"|" "|"{"|"\""|"n"|"e"|"x"|"t"|"\""|":"|" "|"t"|"r"|"u"|"e"|"}"`,
		2:  `"{\""|"pi"|"d\""|":"|=1|=2|"} "|"}i"|"nv"|"al"|"id"|"{ "|"{\""|"ne"|"xt"|"\":"|" t"|"ru"|"e}"`,
		3:  `"{\"p"|"id\""|":"|=12|"} }"|"inv"|"ali"|"d{ "|"{\"n"|"ext"|"\": "|"tru"|"e}"`,
		4:  `"{\"pi"|"d\":"|=12|"} }i"|"nval"|"id{ "|"{\"ne"|"xt\":"|" "|=true|"}"`,
		5:  `"{"|=pid|":"|=12|"} }in"|"valid"|"{ {\"n"|"ext\":"|=true|"}"`,
		6:  `"{"|=pid|":"|=12|"} }inv"|"alid{ "|"{"|=next|":"|=true|"}"`,
		7:  `"{"|=pid|":"|=12|"} }inva"|"lid{ {"|=next|":"|=true|"}"`,
		8:  `"{"|=pid|":"|=12|"} }inval"|"id{ {"|=next|":"|=true|"}"`,
		9:  `"{"|=pid|":"|=12|"} }invali"|"d{ {"|=next|":"|=true|"}"`,
		10: `"{"|=pid|":"|=12|"} }invalid"|"{ {"|=next|":"|=true|"}"`,
		11: `pid=12|" }invalid{ "|"{"|=next|":"|=true|"}"`,
		12: `pid=12|" }invalid{ {"|=next|":"|=true|"}"`,
		14: `pid=12|" }invalid{ "|next=true`,
		15: `pid=12|" }invalid{"|next=true`,
	}

	exp := "notset"
	for readingLimit := 1; readingLimit < len(data); readingLimit++ {
		if a, ok := expectations[readingLimit]; ok {
			exp = a
		}
		for readingChankSize := 1; readingChankSize < len(data); readingChankSize++ {
			readingLimit := readingLimit
			readingChankSize := readingChankSize
			exp := exp
			data := data
			t.Run(fmt.Sprintf("synthetic_single_%d_%d", readingLimit, readingChankSize), func(t *testing.T) {
				t.Parallel()
				buf := &splitReader{data: []byte(data), size: readingChankSize}
				f, g, c := collector(t)
				err := slogtotext.Read(buf, f, g, readingLimit)
				require.NoError(t, err)
				assert.Equal(t, exp, c())
			})
		}
	}

	data = `{
  "pid": 12
}
}invali
d{`
	expectations = map[int]string{
		1:  `"{"|""|" "|" "|"\""|"p"|"i"|"d"|"\""|":"|" "|=1|=2|""|"}"|""|"}"|"i"|"n"|"v"|"a"|"l"|"i"|""|"d"|"{"`,
		2:  `"{"|"  "|"\"p"|"id"|"\":"|=1|=2|""|"}"|"}i"|"nv"|"al"|"i"|"d{"`,
		3:  `"{"|"  \""|"pid"|"\":"|=12|""|"}"|"}in"|"val"|"i"|"d{"`,
		4:  `"{"|"  \"p"|"id\":"|=12|""|"}"|"}inv"|"ali"|"d{"`,
		5:  `"{"|"  "|=pid|":"|=12|""|"}"|"}inva"|"li"|"d{"`,
		6:  `"{"|" "|=pid|":"|=12|""|"}"|"}inval"|"i"|"d{"`,
		7:  `"{"|=pid|":"|=12|""|"}"|"}invali"|""|"d{"`,
		8:  `"{"|=pid|":"|=12|""|"}"|"}invali"|"d{"`,
		15: `pid=12|""|"}invali"|"d{"`,
	}

	exp = "notset"
	for readingLimit := 1; readingLimit < len(data); readingLimit++ {
		if a, ok := expectations[readingLimit]; ok {
			exp = a
		}
		for readingChankSize := 1; readingChankSize < len(data); readingChankSize++ {
			readingLimit := readingLimit
			readingChankSize := readingChankSize
			exp := exp
			data := data
			t.Run(fmt.Sprintf("synthetic_multi_%d_%d", readingLimit, readingChankSize), func(t *testing.T) {
				t.Parallel()
				buf := &splitReader{data: []byte(data), size: readingChankSize}
				f, g, c := collector(t)
				err := slogtotext.Read(buf, f, g, readingLimit)
				require.NoError(t, err)
				assert.Equal(t, exp, c())
			})
		}
	}
}
