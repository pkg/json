package json_test

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/pkg/json"
)

func ExampleScanner_Next() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	sc := json.NewScanner(strings.NewReader(input))
	for {
		tok := sc.Next()
		if len(tok) < 1 {
			break
		}
		fmt.Printf("%s\n", tok)
	}
	if err := sc.Error(); err != nil && err != io.EOF {
		log.Fatal(err)
	}

	// Output:
	// {
	// "a"
	// :
	// 1
	// ,
	// "b"
	// :
	// 123.456
	// ,
	// "c"
	// :
	// [
	// null
	// ]
	// }
}

func ExampleDecoder_Token() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	dec := json.NewDecoder(strings.NewReader(input))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", tok)
	}

	// Output:
	// {
	// a
	// 1
	// b
	// 123.456
	// c
	// [
	// <nil>
	// ]
	// }
}

func ExampleDecoder_NextToken() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	dec := json.NewDecoder(strings.NewReader(input))
	for {
		tok, err := dec.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", tok)
	}

	// Output:
	// {
	// "a"
	// 1
	// "b"
	// 123.456
	// "c"
	// [
	// null
	// ]
	// }
}
func ExampleDecoder_Decode() {
	input := `{"a": 1,"b": 123.456, "c": [null]}`
	dec := json.NewDecoder(strings.NewReader(input))
	var i interface{}
	err := dec.Decode(&i)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", i)

	// Output: map[a:1 b:123.456 c:[<nil>]]
}
