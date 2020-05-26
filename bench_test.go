package json

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var buf [8 << 10]byte

func BenchmarkScanner(b *testing.B) {
	tests := []string{
		// from https://github.com/miloyip/nativejson-benchmark
		"canada.json.gz",
		"citm_catalog.json.gz",
		"twitter.json.gz",
		"code.json.gz",
	}

	for _, tc := range tests {

		f, err := os.Open(filepath.Join("testdata", tc))
		check(b, err)
		defer f.Close()
		gz, err := gzip.NewReader(f)
		check(b, err)
		buf, err := ioutil.ReadAll(gz)
		check(b, err)

		b.Run(tc, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sc := &Scanner{
					r: bytes.NewReader(buf),
					buffer: buffer{
						buf: buf[:],
					},
				}
				for len(sc.Next()) > 0 {

				}
			}
		})
	}
}

func BenchmarkDecoder(b *testing.B) {
	tests := []struct {
		path   string
		tokens int
	}{
		// from https://github.com/miloyip/nativejson-benchmark
		{"canada.json.gz", 223236},
		{"citm_catalog.json.gz", 85035},
		{"twitter.json.gz", 29573},
		{"code.json.gz", 217707},
	}

	for _, tc := range tests {

		f, err := os.Open(filepath.Join("testdata", tc.path))
		check(b, err)
		defer f.Close()
		gz, err := gzip.NewReader(f)
		check(b, err)
		buf, err := ioutil.ReadAll(gz)
		check(b, err)

		b.Run("pkgjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := NewDecoder(bytes.NewReader(buf))
				n := 0
				for {
					_, err := dec.Token()
					if err == io.EOF {
						break
					}
					check(b, err)
					//b.Logf("n: %v, token: %q", n, tok)
					n++
				}
				if n != tc.tokens {
					b.Fatalf("expected %v tokens, got %v", tc.tokens, n)
				}
			}
		})

		b.Run("encodingjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dec := json.NewDecoder(bytes.NewReader(buf))
				n := 0
				for {
					_, err := dec.Token()
					if err == io.EOF {
						break
					}
					check(b, err)
					n++
				}
				if n != tc.tokens {
					b.Fatalf("expected %v tokens, got %v", tc.tokens, n)
				}
			}
		})
	}
}

func check(tb testing.TB, err error) {
	if err != nil {
		tb.Helper()
		tb.Fatal(err)
	}
}
