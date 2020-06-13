package json

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDecoderToken(t *testing.T) {
	tests := []struct {
		json   string
		tokens []string
	}{
		{json: `"a"`, tokens: []string{`"a"`}},
		{json: `1`, tokens: []string{`1`}},
		{json: `{}`, tokens: []string{`{`, `}`}},
		{json: `[]`, tokens: []string{`[`, `]`}},
		{json: `[[[[[[{"true":true}]]]]]]`, tokens: []string{`[`, `[`, `[`, `[`, `[`, `[`, `{`, `"true"`, `true`, `}`, `]`, `]`, `]`, `]`, `]`, `]`}},
		{json: `[{}, {}]`, tokens: []string{`[`, `{`, `}`, `{`, `}`, `]`}},
		{json: `{"a": 0}`, tokens: []string{`{`, `"a"`, `0`, `}`}},
		{json: `{"a": []}`, tokens: []string{`{`, `"a"`, `[`, `]`, `}`}},
		{json: `{"a":{}, "b":{}}`, tokens: []string{`{`, `"a"`, `{`, `}`, `"b"`, `{`, `}`, `}`}},
		{json: `[10]`, tokens: []string{`[`, `10`, `]`}},
		{json: `""`, tokens: []string{`""`}},
		{json: `[{}]`, tokens: []string{`[`, `{`, `}`, `]`}},
		{json: `[{"a": [{}]}]`, tokens: []string{`[`, `{`, `"a"`, `[`, `{`, `}`, `]`, `}`, `]`}},
		{json: `[{"a": 1,"b": 123.456, "c": null, "d": [1, -2, "three", true, false, ""]}]`,
			tokens: []string{`[`,
				`{`,
				`"a"`, `1`,
				`"b"`, `123.456`,
				`"c"`, `null`,
				`"d"`, `[`,
				`1`, `-2`, `"three"`, `true`, `false`, `""`,
				`]`,
				`}`,
				`]`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			dec := NewDecoder(&SmallReader{r: strings.NewReader(tc.json)})
			for n, want := range tc.tokens {
				got, err := dec.Token()
				if string(got) != want {
					t.Fatalf("%v: expected: %q, got: %q, %v", n+1, want, string(got), err)
				}
				t.Logf("token: %q, stack: %v", got, dec.stack)
			}
			last, err := dec.Token()
			if len(last) > 0 {
				t.Fatalf("expected: %q, got: %q, %v", "", string(last), err)
			}
			if err != io.EOF {
				t.Fatalf("expected: %q, got: %q, %v", "", string(last), err)
			}
		})
	}
}

func TestDecoderInvalidJSON(t *testing.T) {
	tests := []struct {
		json string
	}{
		{json: `[`},
		{json: `{"":2`},
		{json: `[[[[]]]`},
		{json: `{"`},
		{json: `{"":` + "\n" + `}`},
		{json: `{{"key": 1}: 2}}`},
		{json: `{1: 1}`},
		// {json: `"\6"`},
		{json: `[[],[], [[]],�[[]]]`},
		{json: `+`},
		{json: `,`},
		// {json: `00`},
		// {json: `1a`},
		{json: `1.e1`},
		{json: `{"a":"b":"c"}`},
		{json: `{"test"::"input"}`},
		{json: `e1`},
		{json: `-.1e-1`},
		{json: `123.`},
		{json: `--123`},
		{json: `.1`},
		{json: `0.1e`},
		// fuzz testing
		// {json: "\"\x00outC: .| >\x185\x014\x80\x00\x01n" +
		//	"E4255425067\x014\x80\x00\x01.242" +
		//	"55425.E420679586036\xef" +
		//	"\xbf9586036�\""},
	}

	for _, tc := range tests {
		t.Run(tc.json, func(t *testing.T) {
			dec := NewDecoder(&SmallReader{r: strings.NewReader(tc.json)})
			var err error
			for {
				_, err = dec.Token()
				if err != nil {
					break
				}
			}
			if err == io.EOF {
				t.Fatalf("expected err, got: %v", err)
			}
		})
	}
}

func TestDecoderDecode(t *testing.T) {

	assert := func(v interface{}, want interface{}) {
		t.Helper()
		got := reflect.ValueOf(v).Interface()
		if !reflect.DeepEqual(want, got) {
			t.Errorf("expected: %v, got: %v", want, got)
		}
	}

	decode := func(input string, v interface{}) {
		dec := NewDecoder(strings.NewReader(input))
		err := dec.Decode(v)
		if err != nil {
			t.Helper()
			t.Errorf("decode %q: %v", input, err)
		}
	}

	var b bool
	decode("true", &b)
	assert(b, true)

	decode("false", &b)
	assert(b, false)

	var bi interface{} = false
	decode("true", &bi)
	assert(bi, true)

	decode("false", &bi)
	assert(bi, false)

	var p = new(int)
	decode("null", &p)
	assert(p, (*int)(nil))

	var m = make(map[int]string)
	decode("null", &m)
	assert(m, (map[int]string)(nil))

	var sl = []string{"a", "b"}
	decode("null", &sl)
	assert(sl, ([]string)(nil))

	var fi interface{}
	decode("3", &fi)
	assert(fi, 3.0)

	var f64 float64
	decode("1", &f64)
	assert(f64, 1.0)

	var f32 float32
	decode("1", &f32)
	assert(f32, float32(1.0))

	return
	var i int
	decode("1", &i)
	assert(i, 1)
}

/*
func TestBitvecPushPop(t *testing.T) {
	var bv bitvec
	for i := 0; i < len(bv.val); i++ {
		v := uint64(i) % 1
		for j := 0; j < i; j++ {
			bv.push(v)
		}
		for j := 0; j < i; j++ {
			if got := bv.pop(); got != (v == 1) {
				t.Fatalf("depth %v: expected: %v, got: %v", j, v, got)
			}
		}
	}
}
*/
