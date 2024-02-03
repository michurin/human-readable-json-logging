//nolint:lll // it's ok for tests
package slogtotext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTXXJson(t *testing.T) {
	data := `{"a":true, "b":null, "c": 1, "d": "xxx", "e": "[1,\"{\\\"p\\\":55}\",3]", "f": false}`
	res := tXXJson(data)
	assert.Equal(t, "{\x1b[93;1ma\x1b[0m:\x1b[92mT\x1b[0m,\x1b[93;1mb\x1b[0m:\x1b[95mN\x1b[0m,\x1b[93;1mc\x1b[0m:\x1b[95m1\x1b[0m,\x1b[93;1md\x1b[0m:\x1b[35mxxx\x1b[0m,\x1b[93;1me\x1b[0m:[\x1b[95m1\x1b[0m,{\x1b[93;1mp\x1b[0m:\x1b[95m55\x1b[0m},\x1b[95m3\x1b[0m],\x1b[93;1mf\x1b[0m:\x1b[91mF\x1b[0m}", res)
}
