package json

import (
	"testing"
)

func BenchmarkCountWhitespace(b *testing.B) {
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
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
			if whitespace[c] {
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
