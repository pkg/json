package json

import (
	"io"
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
		{in: `-1234567.8e+90`, tokens: []string{`-1234567.8e+90`}},
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
	testParseString(t, `""`, `""`)
	testParseString(t, `"" `, `""`)
	testParseString(t, `"\""`, `"\""`)
	testParseString(t, `"\\\\\\\\\6"`, `"\\\\\\\\\6"`)
	testParseString(t, `"\6"`, `"\6"`)
}

func testParseString(t *testing.T, json, want string) {
	t.Helper()
	r := strings.NewReader(json)
	scanner := NewScanner(r)
	got := scanner.Next()
	if string(got) != want {
		t.Fatalf("expected: %q, got: %q", want, got)
	}
}

func TestParseNumber(t *testing.T) {
	testParseNumber(t, `1`)
	// testParseNumber(t, `0000001`)
	testParseNumber(t, `12.0004`)
	testParseNumber(t, `1.7734`)
	testParseNumber(t, `15`)
	testParseNumber(t, `-42`)
	testParseNumber(t, `-1.7734`)
	testParseNumber(t, `1.0e+28`)
	testParseNumber(t, `-1.0e+28`)
	testParseNumber(t, `1.0e-28`)
	testParseNumber(t, `-1.0e-28`)
	testParseNumber(t, `-18.3872`)
	testParseNumber(t, `-2.1`)
	testParseNumber(t, `-1234567.891011121314`)
}

func testParseNumber(t *testing.T, tc string) {
	t.Helper()
	r := strings.NewReader(tc)
	scanner := NewScanner(r)
	got := scanner.Next()
	if string(got) != tc {
		t.Fatalf("expected: %q, got: %q", tc, got)
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
			b.SetBytes(r.Size())
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.Seek(0, 0)
				scanner := &Scanner{
					br: byteReader{
						data: buf[:0],
						r:    r,
					},
				}
				scanner.br.extend()
				n := scanner.parseNumber(scanner.br.window()[0])
				if n != len(tc) {
					b.Fatalf("expected: %v, got: %v", len(tc), n)
				}
			}
		})
	}
}

func TestScanner(t *testing.T) {
	testScanner(t, 1)
	testScanner(t, 8)
	testScanner(t, 64)
	testScanner(t, 256)
	testScanner(t, 1<<10)
	testScanner(t, 8<<10)
	testScanner(t, 1<<20)
}

func testScanner(t *testing.T, sz int) {
	t.Helper()
	buf := make([]byte, sz)
	for _, tc := range inputs {
		r := fixture(t, tc.path)
		t.Run(tc.path, func(t *testing.T) {
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
				t.Fatalf("expected %v tokens, got %v", tc.alltokens, n)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
