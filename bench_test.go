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

var _buf [8 << 10]byte

var inputs = []struct {
	path   string
	tokens int
}{
	// from https://github.com/miloyip/nativejson-benchmark
	{"canada.json.gz", 223236},
	{"citm_catalog.json.gz", 85035},
	{"twitter.json.gz", 29573},
	{"code.json.gz", 217707},
}

func BenchmarkScanner(b *testing.B) {
	for _, tc := range inputs {

		f, err := os.Open(filepath.Join("testdata", tc.path))
		check(b, err)
		defer f.Close()
		gz, err := gzip.NewReader(f)
		check(b, err)
		buf, err := ioutil.ReadAll(gz)
		check(b, err)
		r := bytes.NewReader(buf)

		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				sc := &Scanner{
					r: r,
					buffer: buffer{
						buf: _buf[:],
					},
				}
				for len(sc.Next()) > 0 {

				}
			}
		})
	}
}

func BenchmarkDecoder(b *testing.B) {
	for _, tc := range inputs {

		f, err := os.Open(filepath.Join("testdata", tc.path))
		check(b, err)
		defer f.Close()
		gz, err := gzip.NewReader(f)
		check(b, err)
		buf, err := ioutil.ReadAll(gz)
		check(b, err)

		r := bytes.NewReader(buf)

		b.Run("pkgjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				dec := NewDecoderBuffer(r, _buf[:])
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
				r.Seek(0, 0)
				dec := json.NewDecoder(r)
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

func BenchmarkUnbufferedDecoder(b *testing.B) {
	for _, tc := range inputs {
		b.Run("pkgjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				f, err := os.Open(filepath.Join("testdata", tc.path))
				check(b, err)
				gz, err := gzip.NewReader(f)
				check(b, err)
				dec := NewDecoderBuffer(gz, _buf[:])
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
				f.Close()
				if n != tc.tokens {
					b.Fatalf("expected %v tokens, got %v", tc.tokens, n)
				}
			}
		})

		b.Run("encodingjson/"+tc.path, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f, err := os.Open(filepath.Join("testdata", tc.path))
				check(b, err)
				gz, err := gzip.NewReader(f)
				check(b, err)
				dec := json.NewDecoder(gz)
				n := 0
				for {
					_, err := dec.Token()
					if err == io.EOF {
						break
					}
					check(b, err)
					n++
				}
				f.Close()
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
