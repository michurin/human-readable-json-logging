package slogtotext

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode"
)

const (
	allKey         = "ALL"
	rawInputKey    = "RAW_INPUT"
	invalidLineKey = "TEXT"
	binaryKey      = "BINARY"
)

func tryToParse(b []byte) (any, bool) {
	if !json.Valid(b) {
		return nil, false
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	data := any(nil)
	err := d.Decode(&data)
	if err != nil {
		return nil, false
	}
	return data, true
}

func Read(input io.Reader, out func([]Pair) error, outStr func([]Pair) error, maxCap int) error { // TODO wrap errors
	sc := bufio.NewScanner(input)
	buf := make([]byte, maxCap)
	sc.Buffer(buf, maxCap)
	for sc.Scan() {
		data, ok := tryToParse(sc.Bytes())
		if ok {
			rec := flat(data)
			rec = append(rec, Pair{rawInputKey, sc.Text()})
			err := out(rec)
			if err != nil {
				return err
			}
		} else {
			s := sc.Text()
			x := ""
			if strings.IndexFunc(s, unicode.IsControl) >= 0 { // strings.ContainsFunc shows up in go go1.21
				x = "yes"
			}
			err := outStr([]Pair{
				{K: invalidLineKey, V: s},
				{K: binaryKey, V: x},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
