package slogtotext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTXJson(t *testing.T) {
	data := `{"a":true, "b":null, "c": 1, "d": "xxx", "e": "[1,\"{\\\"p\\\":55}\",3]"}`
	res := tXJson(data)
	assert.Equal(t, `{"a":true,"b":null,"c":1,"d":"xxx","e":[1,{"p":55},3]}`, res)
}
