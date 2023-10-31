package json

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var inputs = []struct {
	path       string
	tokens     int // decoded tokens
	alltokens  int // raw tokens, includes : and ,
	whitespace int // number of whitespace chars
}{
	// from https://github.com/miloyip/nativejson-benchmark
	{"canada", 223236, 334373, 33},
	{"citm_catalog", 85035, 135990, 1227563},
	{"twitter", 29573, 55263, 167931},
	{"code", 217707, 396293, 3},

	// from https://raw.githubusercontent.com/mailru/easyjson/master/benchmark/example.json
	{"example", 710, 1297, 4246},

	// from https://github.com/ultrajson/ultrajson/blob/master/tests/sample.json
	{"sample", 5276, 8677, 518549},
}

func BenchmarkScanner(b *testing.B) {
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				sc := &Scanner{
					br: byteReader{
						data: buf[:0],
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
	b.Skip()
	sizes := []int{16, 64, 256, 512, 1 << 10, 2 << 10, 4 << 10, 8 << 10, 16 << 10, 64 << 10, 1 << 20}
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			for _, sz := range sizes {
				buf := make([]byte, sz)
				b.Run(strconv.Itoa(sz), func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(r.Size())
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
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run("pkgjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				dec := NewDecoderBuffer(r, buf[:])
				var i interface{}
				err := dec.Decode(&i)
				check(b, err)
			}
		})
		b.Run("encodingjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
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
	var buf [8 << 10]byte
	in := `{"a": 97, "b": 98, "c": 99, "d": 100, "e": 101, "f": 102, "g": 103 }`
	r := strings.NewReader(in)
	b.Run("pkgjson", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(r.Size())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Seek(0, 0)
			dec := NewDecoderBuffer(r, buf[:])
			m := make(map[string]int)
			err := dec.Decode(&m)
			check(b, err)
		}
	})
	b.Run("encodingjson", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(r.Size())
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
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run("pkgjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				dec := NewDecoderBuffer(r, buf[:])
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
			b.SetBytes(r.Size())
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
	var buf [8 << 10]byte
	for _, tc := range inputs {
		r := fixture(b, tc.path)
		b.Run("pkgjson/"+tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(r.Size())
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				dec := NewDecoderBuffer(r, buf[:])
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
			b.SetBytes(r.Size())
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

// fuxture returns a *bytes.Reader for the contents of path.
func fixture(tb testing.TB, path string) *bytes.Reader {
	f, err := os.Open(filepath.Join("testdata", path+".json.gz"))
	check(tb, err)
	defer f.Close()
	gz, err := gzip.NewReader(f)
	check(tb, err)
	buf, err := io.ReadAll(gz)
	check(tb, err)
	return bytes.NewReader(buf)
}

func check(tb testing.TB, err error) {
	if err != nil {
		tb.Helper()
		tb.Fatal(err)
	}
}
