// +build gofuzz

package json

import (
	"bytes"
	"io"
)

var buf [8 << 10]byte

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

	dec := NewDecoderBuffer(bytes.NewReader(data), buf[:])
	for {
		_, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return 1
			}
			return -1
		}
		return 0
	}
}
