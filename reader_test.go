package json

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCountWhitespace(b *testing.B) {
	var buf [8 << 10]byte
	for _, tc := range inputs {

		f, err := os.Open(filepath.Join("testdata", tc.path))
		check(b, err)
		defer f.Close()
		gz, err := gzip.NewReader(f)
		check(b, err)
		data, err := ioutil.ReadAll(gz)
		check(b, err)
		r := bytes.NewReader(data)

		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				br := byteReader{
					data: buf[:0],
					r:    r,
				}
				got := countWhitespace(&br)
				if got != tc.whitespace {
					b.Fatalf("expected: %v, got: %v", tc.whitespace, got)
				}
			}
		})
	}
}

func countWhitespace(br *byteReader) int {
	n := 0
	w := br.window()
	for {
		for _, c := range w {
			if isWhitespace(c) {
				n++
			}
		}
		br.release(len(w))
		if br.extend() == 0 {
			return n
		}
		w = br.window()
	}
}
