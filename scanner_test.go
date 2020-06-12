package json

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type SmallReader struct {
	r io.Reader
	n int
}

func (sm *SmallReader) next() int {
	sm.n = (sm.n + 3) % 5
	if sm.n < 1 {
		sm.n++
	}
	return sm.n
}

func (sm *SmallReader) Read(buf []byte) (int, error) {
	return sm.r.Read(buf[:min(sm.next(), len(buf))])
}

func TestScannerNext(t *testing.T) {
	tests := []struct {
		in     string
		tokens []string
	}{
		{in: `""`, tokens: []string{`""`}},
		{in: `"a"`, tokens: []string{`"a"`}},
		{in: ` "a" `, tokens: []string{`"a"`}},
		{in: `"\""`, tokens: []string{`"\""`}},
		{in: `1`, tokens: []string{`1`}},
		{in: `{}`, tokens: []string{`{`, `}`}},
		{in: `[]`, tokens: []string{`[`, `]`}},
		{in: `[{}, {}]`, tokens: []string{`[`, `{`, `}`, `,`, `{`, `}`, `]`}},
		{in: `{"a": 0}`, tokens: []string{`{`, `"a"`, `:`, `0`, `}`}},
		{in: `{"a": []}`, tokens: []string{`{`, `"a"`, `:`, `[`, `]`, `}`}},
		{in: `[10]`, tokens: []string{`[`, `10`, `]`}},
		{in: `[{"a": 1,"b": 123.456, "c": null, "d": [1, -2, "three", true, false, ""]}]`,
			tokens: []string{`[`,
				`{`,
				`"a"`, `:`, `1`, `,`,
				`"b"`, `:`, `123.456`, `,`,
				`"c"`, `:`, `null`, `,`,
				`"d"`, `:`, `[`,
				`1`, `,`, `-2`, `,`, `"three"`, `,`, `true`, `,`, `false`, `,`, `""`,
				`]`,
				`}`,
				`]`,
			},
		},
		{in: `{"x": "va\\\\ue", "y": "value y"}`, tokens: []string{
			`{`, `"x"`, `:`, `"va\\\\ue"`, `,`, `"y"`, `:`, `"value y"`, `}`,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			scanner := NewScanner(&SmallReader{r: strings.NewReader(tc.in)})
			for n, want := range tc.tokens {
				got := scanner.Next()
				if string(got) != want {
					t.Fatalf("%v: expected: %v, got: %v", n+1, want, string(got))
				}
			}
			last := scanner.Next()
			if len(last) > 0 {
				t.Fatalf("expected: %q, got: %q", "", string(last))
			}
			if err := scanner.Error(); err != io.EOF {
				t.Fatalf("expected: %v, got: %v", io.EOF, err)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		json string
		want string
	}{
		{`""`, `""`},
		{`"" `, `""`},
		{`"\""`, `"\""`},
		{`"\\\\\\\\\6"`, `"\\\\\\\\\6"`},
		{`"\6"`, `"\6"`},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			r := strings.NewReader(tc.json)
			scanner := &Scanner{
				r: r,
			}
			scanner.extend(0) // consume reader
			n := scanner.parseString()
			if n != len(tc.want) {
				t.Fatalf("expected: %v, got: %v", len(tc.want), n)
			}
			got := scanner.buffer.window()[:n]
			if string(got) != tc.want {
				t.Fatalf("expected: %q, got: %q", tc.want, got)
			}
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []string{
		`1`,
		// `00`,
		`12.0004`,
		`1.7734`,
		`15`,
		`-42`,
		`-1.7734`,
		`1.0e+28`,
		`-1.0e+28`,
		`1.0e-28`,
		`-1.0e-28`,
		`-18.3872`,
		`-2.1`,
		`-1234567.891011121314`,
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			r := strings.NewReader(tc)
			scanner := &Scanner{
				r: r,
			}
			scanner.extend(0) // consume reader
			n := scanner.parseNumber()
			if n != len(tc) {
				t.Fatalf("expected: %v, got: %v", len(tc), n)
			}
			got := scanner.buffer.window()[:n]
			if string(got) != tc {
				t.Fatalf("expected: %q, got: %q", tc, got)
			}
		})
	}
}

func BenchmarkParseNumber(b *testing.B) {
	tests := []string{
		`1`,
		`12.0004`,
		`1.7734`,
		`15`,
		`-42`,
		`-1.7734`,
		`1.0e+28`,
		`-1.0e+28`,
		`1.0e-28`,
		`-1.0e-28`,
		`-18.3872`,
		`-2.1`,
		`-1234567.891011121314`,
	}
	var buf [4 << 10]byte

	for _, tc := range tests {
		r := strings.NewReader(tc)
		b.Run(tc, func(b *testing.B) {
			b.SetBytes(int64(len(tc)))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				scanner := &Scanner{
					r: r,
					buffer: buffer{
						buf: buf[:0],
					},
				}
				n := scanner.parseNumber()
				if n != len(tc) {
					b.Fatalf("failed")
				}
			}
		})
	}
}

func TestScanner(t *testing.T) {
	for _, tc := range inputs {

		f, err := os.Open(filepath.Join("testdata", tc.path))
		check(t, err)
		defer f.Close()
		gz, err := gzip.NewReader(f)
		check(t, err)
		r := &SmallReader{r: gz}

		t.Run(tc.path, func(t *testing.T) {
			sc := &Scanner{
				r: r,
				buffer: buffer{
					buf: _buf[:0],
				},
			}
			n := 0
			for len(sc.Next()) > 0 {
				n++
			}
			if n != tc.alltokens {
				t.Fatalf("expected %v tokens, got %v", tc.alltokens, n)
			}
		})
	}
}
