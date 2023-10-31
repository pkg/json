//go:build gofuzz
// +build gofuzz

package json

import (
	"bytes"
	"io"
)

func Fuzz(data []byte) int {
	sc := NewScanner(bytes.NewReader(data))
	for {
		tok := sc.Next()
		if len(tok) < 1 {
			if sc.Error() != nil {
				return -1
			}
			break
		}
	}

	dec := NewDecoder(bytes.NewReader(data))
	for {
		_, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return -1
		}
		return 0
	}
	var i interface{}
	dec = NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&i)
	if err != nil {
		return -1
	}
	return 1
}
