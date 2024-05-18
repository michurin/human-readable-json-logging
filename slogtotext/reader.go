package slogtotext

import (
	"bytes"
	"encoding/json"
	"io"
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
	// This implementation is slightly naive (extra reading/coping prone) and mimics bufio.Scan.
	// However, we do not use bufio.Scan because it consider too long taken as error (ErrTooLong).
	// Considering of this error makes code too ugly because
	// in our case we do not consider it as error, however it is special case.
	// The room for refactoring.
	buf := make([]byte, maxCap)
	line := []byte(nil)
	bufEnd := 0
	for {
		n, err := input.Read(buf[bufEnd:])
		noMoreData := n == 0 && err != nil // it is the last data
		if noMoreData && bufEnd == 0 {
			// Read can return n > 0 and EOF, however it must return 0 and EOF next time
			// And we give the split function a chance, like bufio.Scon does (check bufEnd == 0)
			break
		}
		bufEnd += n
		for bufEnd > 0 {
			s := bytes.IndexByte(buf[:bufEnd], '\n')
			if s < 0 {
				if bufEnd == maxCap || noMoreData { // consider full buffer
					line = buf[:bufEnd]
					buf = make([]byte, maxCap)
					bufEnd = 0 // will cause end of iterations
				} else {
					break // the buffer is not full, however, it is not a complete line
				}
			} else {
				line = buf[:s]
				x := make([]byte, maxCap)
				copy(x, buf[s+1:])
				buf = x
				bufEnd -= s + 1
			}
			data, ok := tryToParse(line)
			if ok {
				rec := flat(data)
				rec = append(rec, Pair{rawInputKey, string(line)})
				err := out(rec)
				if err != nil {
					return err
				}
			} else {
				x := ""
				if bytes.IndexFunc(line, unicode.IsControl) >= 0 { // bytes.ContainsFunc shows up in go go1.21
					x = "yes"
				}
				err := outStr([]Pair{
					{K: invalidLineKey, V: string(line)},
					{K: binaryKey, V: x},
				})
				if err != nil {
					return err
				}
			}
		}
		if noMoreData {
			break
		}
	}
	return nil
}
