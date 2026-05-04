package json

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDecoderNextToken(t *testing.T) {
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
				got, err := dec.NextToken()
				if string(got) != want {
					t.Fatalf("%v: expected: %q, got: %q, %v", n+1, want, string(got), err)
				}
				t.Logf("token: %q, stack: %v", got, dec.stack)
			}
			last, err := dec.NextToken()
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

	assert := func(t *testing.T, v interface{}, want interface{}) {
		t.Helper()
		got := reflect.ValueOf(v).Interface()
		if !reflect.DeepEqual(want, got) {
			t.Errorf("expected: %v, got: %v", want, got)
		}
	}

	decode := func(t *testing.T, input string, v interface{}) {
		t.Helper()
		dec := NewDecoder(strings.NewReader(input))
		err := dec.Decode(v)
		if err != nil {
			t.Errorf("decode %q: %v", input, err)
		}
	}

	t.Run("bool true", func(t *testing.T) {
		var b bool
		decode(t, "true", &b)
		assert(t, b, true)
	})

	t.Run("bool false", func(t *testing.T) {
		var b bool
		decode(t, "false", &b)
		assert(t, b, false)
	})

	t.Run("bool interface true", func(t *testing.T) {
		var bi interface{} = false
		decode(t, "true", &bi)
		assert(t, bi, true)
	})

	t.Run("bool interface false", func(t *testing.T) {
		var bi interface{} = true
		decode(t, "false", &bi)
		assert(t, bi, false)
	})

	t.Run("null pointer", func(t *testing.T) {
		var p = new(int)
		decode(t, "null", &p)
		assert(t, p, (*int)(nil))
	})

	t.Run("null map", func(t *testing.T) {
		var m = make(map[int]string)
		decode(t, "null", &m)
		assert(t, m, (map[int]string)(nil))
	})

	t.Run("null slice", func(t *testing.T) {
		var sl = []string{"a", "b"}
		decode(t, "null", &sl)
		assert(t, sl, ([]string)(nil))
	})

	t.Run("float64 interface", func(t *testing.T) {
		var fi interface{}
		decode(t, "3", &fi)
		assert(t, fi, 3.0)
	})

	t.Run("float64", func(t *testing.T) {
		var f64 float64
		decode(t, "1", &f64)
		assert(t, f64, 1.0)
	})

	t.Run("float32", func(t *testing.T) {
		var f32 float32
		decode(t, "1", &f32)
		assert(t, f32, float32(1.0))
	})

	t.Run("int", func(t *testing.T) {
		var i int
		decode(t, "1", &i)
		assert(t, i, 1)
	})

	t.Run("int64", func(t *testing.T) {
		var i64 int64
		decode(t, "-1", &i64)
		assert(t, i64, int64(-1))
	})

	t.Run("uint", func(t *testing.T) {
		var u uint
		decode(t, "1", &u)
		assert(t, u, uint(1))
	})

	t.Run("empty object", func(t *testing.T) {
		var a interface{}
		decode(t, "{}", &a)
		assert(t, a, map[string]interface{}{})
	})

	t.Run("nested object", func(t *testing.T) {
		var a interface{}
		decode(t, `{"a": 1, "b": {"c": 2}}`, &a)
		assert(t, a, map[string]interface{}{
			"a": float64(1),
			"b": map[string]interface{}{
				"c": float64(2),
			},
		})
	})

	t.Run("nested array of objects", func(t *testing.T) {
		var a interface{}
		decode(t, `[{"a": [{}]}]`, &a)
		assert(t, a, []interface{}{
			map[string]interface{}{
				"a": []interface{}{
					map[string]interface{}{},
				},
			},
		})
	})

	t.Run("object key with embedded quote", func(t *testing.T) {
		t.Skip("known bug: decoder does not unescape backslashes in object keys")
		var escaped interface{}
		decode(t, `{"a\"b":0}`, &escaped)
		assert(t, escaped, map[string]interface{}{`a"b`: 0.0})
	})

	t.Run("map string string", func(t *testing.T) {
		ms := make(map[string]string)
		decode(t, `{"hello": "world"}`, &ms)
		assert(t, ms, map[string]string{
			"hello": "world",
		})
	})

	t.Run("map string interface", func(t *testing.T) {
		mi := make(map[string]interface{})
		decode(t, `{"a": 1, "b": false, "c":[1, 2.0, "three"]}`, &mi)
		assert(t, mi, map[string]interface{}{
			"a": float64(1),
			"b": false,
			"c": []interface{}{
				float64(1),
				2.0,
				"three",
			},
		})
	})
}
