package slogtotext

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const invalidLineKey = "text"

func decoder(a, b io.Reader, limit int) *json.Decoder {
	d := json.NewDecoder(io.LimitReader(io.MultiReader(a, b), int64(limit)))
	d.UseNumber()
	return d
}

func Read(input io.Reader, out func([]Pair) error, outStr func([]Pair) error, maxData int) error { // TODO wrap errors
	oneByteBuff := make([]byte, 1)
	waste := make([]byte, 0, maxData)
	data := any(nil)
	dec := decoder(new(bytes.Buffer), input, maxData)
	for {
		err := dec.Decode(&data)
		if err != nil {
			b := dec.Buffered()
			n, err := b.Read(oneByteBuff) // skip only one char and try again
			if err != nil {
				if errors.Is(err, io.EOF) { // real EOF, we can not read even one byte
					break
				}
				return err
			}
			if n != 1 {
				return fmt.Errorf("not 1 byte is read: %d", n) // TODO how it possible?
			}
			if oneByteBuff[0] == '\n' {
				if len(waste) != 0 { // NL appears right after valid JSON
					err = outStr([]Pair{{K: invalidLineKey, V: string(waste)}})
					if err != nil {
						return err // TODO wrap
					}
					waste = waste[:0]
				}
			} else {
				waste = append(waste, oneByteBuff[0])
				if len(waste) >= maxData {
					err = outStr([]Pair{{K: invalidLineKey, V: string(waste)}})
					if err != nil {
						return err // TODO wrap
					}
					waste = waste[:0]
				}
			}
			dec = decoder(b, input, maxData)
			continue
		}
		if len(waste) > 0 {
			err = outStr([]Pair{{K: invalidLineKey, V: string(waste)}})
			if err != nil {
				return err
			}
			waste = waste[:0]
		}
		err = out(flat(data))
		if err != nil {
			return err
		}
		dec = decoder(dec.Buffered(), input, maxData) // we have to reset limit
		data = nil
	}
	if len(waste) > 0 { // TODO? dup
		err := outStr([]Pair{{K: invalidLineKey, V: string(waste)}})
		if err != nil {
			return err
		}
	}
	return nil
}
