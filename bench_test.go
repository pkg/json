package json

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var _buf [8 << 10]byte

var inputs = []struct {
	path       string
	tokens     int // decoded tokens
	alltokens  int // raw tokens, includes : and ,
	whitespace int // number of whitespace chars
}{
	// from https://github.com/miloyip/nativejson-benchmark
	{"canada.json.gz", 223236, 334373, 33},
	{"citm_catalog.json.gz", 85035, 135990, 1227563},
	{"twitter.json.gz", 29573, 55263, 167931},
	{"code.json.gz", 217707, 396293, 3},

	// from https://raw.githubusercontent.com/mailru/easyjson/master/benchmark/example.json
	{"example.json.gz", 710, 1297, 4246},

	// from https://github.com/ultrajson/ultrajson/blob/master/tests/sample.json
	{"sample.json.gz", 5276, 8677, 518549},
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
					br: byteReader{
						data: _buf[:0],
						r:    r,
					},
				}
				n := 0
				for len(sc.Next()) > 0 {
					n++
				}
				if n != tc.alltokens {
					b.Fatalf("expected %v tokens, got %v", tc.alltokens, n)
				}

			}
		})
	}
}

func BenchmarkBufferSize(b *testing.B) {
	sizes := []int{
		64, 256, 1 << 10, 8 << 10, 1 << 20,
	}

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
			for _, sz := range sizes {
				buf := make([]byte, sz)
				b.Run(fmt.Sprint(sz), func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(buf)))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						r.Seek(0, 0)
						sc := &Scanner{
							br: byteReader{
								data: buf[:0],
								r:    r,
							},
						}
						for len(sc.Next()) > 0 {

						}
					}
				})
			}
		})
	}
}

func BenchmarkDecoderDecodeInterfaceAny(b *testing.B) {
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
				var i interface{}
				err := dec.Decode(&i)
				check(b, err)
			}
		})

		b.Run("encodingjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(buf)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				dec := json.NewDecoder(r)
				var i interface{}
				err := dec.Decode(&i)
				check(b, err)
			}
		})
	}
}

func BenchmarkDecoderDecodeMapInt(b *testing.B) {
	in := `{"a": 97, "b": 98, "c": 99, "d": 100, "e": 101, "f": 102, "g": 103 }`
	r := strings.NewReader(in)
	b.Run("pkgjson", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := NewDecoderBuffer(r, _buf[:])
			m := make(map[string]int)
			err := dec.Decode(&m)
			check(b, err)
		}
	})

	b.Run("encodingjson", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := json.NewDecoder(r)
			m := make(map[string]int)
			err := dec.Decode(&m)
			check(b, err)
		}
	})
}

func BenchmarkDecoderToken(b *testing.B) {
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

func BenchmarkDecoderNextToken(b *testing.B) {
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
					_, err := dec.NextToken()
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

func check(tb testing.TB, err error) {
	if err != nil {
		tb.Helper()
		tb.Fatal(err)
	}
}
