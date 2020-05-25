package json

import (
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	data := `{"a": 1,"b": 123.456, "c": null}`
	s := &source{
		r: strings.NewReader(data),
	}

	check := func(item JSONItem, token byte, expected string) {
		t.Helper()
		if item.token != token {
			t.Fatalf("expected item.token: %v, got: %v\n", token, item.token)
		}
		if got := string(item.data(s)); got != expected {
			t.Fatalf("expected token value: %s, got: %s\n", expected, got)
		}
	}

	parser := &Tokenizer{
		source: s,
	}

	check(parser.Next(), ObjectStart, "{")
	check(parser.Next(), String, "a")
	check(parser.Next(), Colon, ":")
	check(parser.Next(), Number, "1")
	check(parser.Next(), Comma, ",")
	check(parser.Next(), String, "b")
	check(parser.Next(), Colon, ":")
}
