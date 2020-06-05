// package json decodes JSON.
package json

import (
	"fmt"
	"io"
)

// NewDecoder returns a new Decoder for the supplied Reader r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		scanner: &Scanner{
			r: r,
		},
		step: stateValue,
	}
}

// A Decoder decodes JSON values from an input stream.
type Decoder struct {
	scanner *Scanner
	step    func(*Decoder) ([]byte, error)
	stack   []bool
}

// Token returns a []byte referencing the next logical token in the stream.
// The []byte is valid until Token is called again.
// Ad the end of the input stream, Token returns nil, io.EOF.
//
// Token guarantees that the delimiters [ ] { } it returns are properly nested
// and matched: if Token encounters an unexpected delimiter in the input, it
// will return an error.
//
// A valid token begins with one of the following:
//
//    { Object start
//    [ Array start
//    } Object end
//    ] Array End
//    t JSON true
//    f JSON false
//    n JSON null
//    " A string, possibly containing backslash escaped entites.
//    -, 0-9 A number
//
// Commas and colons are elided.
func (d *Decoder) Token() ([]byte, error) {
	return d.step(d)
}

func (d *Decoder) pop() bool {
	d.stack = d.stack[:len(d.stack)-1]
	if len(d.stack) == 0 {
		return false
	}
	return d.stack[len(d.stack)-1]
}

func stateEnd(d *Decoder) ([]byte, error) {
	return nil, io.EOF
}

func stateObjectString(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case len(d.stack) == 0:
			d.step = stateEnd
		case inObj:
			d.step = stateObjectComma
		case !inObj:
			d.step = stateArrayComma
		}
		return tok, nil
	case '"':
		d.step = stateObjectColon
		return tok, nil
	default:
		return nil, fmt.Errorf("stateObjectString: missing string key")
	}
}

func stateObjectColon(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case Colon:
		d.step = stateObjectValue
		return d.Token()
	default:
		return tok, fmt.Errorf("stateObjectColon: expecting colon")
	}
}

func stateObjectValue(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case '{':
		d.step = stateObjectString
		d.stack = append(d.stack, true)
		return tok, nil
	case '[':
		d.step = stateArrayValue
		d.stack = append(d.stack, false)
		return tok, nil
	default:
		d.step = stateObjectComma
		return tok, nil
	}
}

func stateObjectComma(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case len(d.stack) == 0:
			d.step = stateValue
		case inObj:
			d.step = stateObjectComma
		case !inObj:
			d.step = stateArrayComma
		}
		return tok, nil
	case Comma:
		d.step = stateObjectString
		return d.Token()
	default:
		return tok, fmt.Errorf("stateObjectComma: expecting comma")
	}
}

func stateArrayValue(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case '{':
		d.step = stateObjectString
		d.stack = append(d.stack, true)
		return tok, nil
	case '[':
		d.step = stateArrayValue
		d.stack = append(d.stack, false)
		return tok, nil
	case ']':
		inObj := d.pop()
		switch {
		case len(d.stack) == 0:
			d.step = stateEnd
		case inObj:
			d.step = stateObjectComma
		case !inObj:
			d.step = stateArrayComma
		}
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateArrayValue: unexpected comma")
	default:
		d.step = stateArrayComma
		return tok, nil
	}
}

func stateArrayComma(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case ']':
		inObj := d.pop()
		switch {
		case len(d.stack) == 0:
			d.step = stateEnd
		case inObj:
			d.step = stateObjectComma
		case !inObj:
			d.step = stateArrayComma
		}
		return tok, nil
	case Comma:
		d.step = stateArrayValue
		return d.Token()
	default:
		return nil, fmt.Errorf("stateArrayComma: expected comma")
	}
}

func stateValue(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, d.scanner.Error()
	}
	switch tok[0] {
	case '{':
		d.step = stateObjectString
		d.stack = append(d.stack, true)
		return tok, nil
	case '[':
		d.step = stateArrayValue
		d.stack = append(d.stack, false)
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateValue: unexpected comma")
	default:
		d.step = stateEnd
		return tok, nil
	}
}
