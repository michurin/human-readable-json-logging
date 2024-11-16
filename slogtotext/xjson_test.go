package slogtotext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTXJson(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		data := `{"a": true, "b": null, "c": 1, "d": "xxx", "e": "[1,\"{\\\"p\\\":55}\",3]"}`
		res := tXJson(data)
		assert.Equal(t, `{"a":true,"b":null,"c":1,"d":"xxx","e":[1,{"p":55},3]}`, res) //nolint:testifylint
	})
	t.Run("raw", func(t *testing.T) {
		data := `xxx`
		res := tXJson(data)
		assert.Equal(t, `xxx`, res)
	})
}
