// Package json decodes JSON.
package json

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// NewDecoder returns a new Decoder for the supplied Reader r.
func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderBuffer(r, make([]byte, 8192))
}

// NewDecoderBuffer returns a new Decoder for the supplier Reader r, using
// the []byte buf provided for working storage.
func NewDecoderBuffer(r io.Reader, buf []byte) *Decoder {
	return &Decoder{
		scanner: &Scanner{
			r: r,
			buffer: buffer{
				buf: buf[:0],
			},
		},
		step: stateValue,
	}
}

type bitvec struct {
	len int
	val uint64
}

func (bv *bitvec) push(v uint64) {
	bv.val |= v << bv.len
	bv.len++
}

func (bv *bitvec) pop() bool {
	v := bv.val&1<<bv.len != 0
	bv.len--
	return v
}

type stack []bool

func (s *stack) push(v bool) {
	*s = append(*s, v)
}

func (s *stack) pop() bool {
	*s = (*s)[:len(*s)-1]
	if len(*s) == 0 {
		return false
	}
	return (*s)[len(*s)-1]
}

func (s *stack) len() int { return len(*s) }

// A Decoder decodes JSON values from an input stream.
type Decoder struct {
	scanner *Scanner
	step    func(*Decoder) ([]byte, error)
	// bitvec
	stack
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

func stateEnd(d *Decoder) ([]byte, error) {
	return nil, io.EOF
}

func stateObjectString(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
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
		return nil, io.ErrUnexpectedEOF
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
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.step = stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.step = stateArrayValue
		d.push(false)
		return tok, nil
	default:
		d.step = stateObjectComma
		return tok, nil
	}
}

func stateObjectComma(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.step = stateEnd
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
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.step = stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.step = stateArrayValue
		d.push(false)
		return tok, nil
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
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
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
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
		return nil, fmt.Errorf("stateArrayComma: expected comma, %v", d.stack)
	}
}

func stateValue(d *Decoder) ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.step = stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.step = stateArrayValue
		d.push(false)
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateValue: unexpected comma")
	default:
		d.step = stateEnd
		return tok, nil
	}
}

// Decode reads the next JSON-encoded value from its input and stores it
// in the value pointed to by v.
func (d *Decoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	switch {
	case rv.Kind() != reflect.Ptr:
		return fmt.Errorf("non-pointer %v", reflect.TypeOf(v))
	case rv.IsNil():
		return fmt.Errorf("nil")
	default:
		return d.decodeValue(rv.Elem())
	}
}

func (d *Decoder) decodeValue(v reflect.Value) error {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return d.scanner.Error()
	}
	switch tok[0] {
	case True, False:
		value := tok[0] == 't'
		switch v.Kind() {
		case reflect.Bool:
			v.SetBool(value)
		case reflect.Interface:
			if v.NumMethod() > 0 {
				return fmt.Errorf("cannot decode bool into Go value of type %v", v.Type())
			}
			v.Set(reflect.ValueOf(value))
		default:
			return fmt.Errorf("unhandled type: %v", v.Kind())
		}
		return nil
	case Null:
		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice:
			v.Set(reflect.Zero(v.Type()))
			return nil
		default:
			return fmt.Errorf("unhandled type: %v", v.Kind())
		}
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		switch v.Kind() {
		case reflect.Interface:
			if v.NumMethod() > 0 {
				return fmt.Errorf("cannot decode number into Go value of type %v", v.Type())
			}
			f, err := strconv.ParseFloat(string(tok), 64)
			if err != nil {
				return fmt.Errorf("cannot convert %q to float: %v", tok, err)
			}
			v.Set(reflect.ValueOf(f))
		case reflect.Float64:
			f, err := strconv.ParseFloat(string(tok), 64)
			if err != nil {
				return fmt.Errorf("cannot convert %q to float: %v", tok, err)
			}
			v.Set(reflect.ValueOf(f))
		case reflect.Float32:
			f, err := strconv.ParseFloat(string(tok), 64)
			if err != nil {
				return fmt.Errorf("cannot convert %q to float: %v", tok, err)
			}
			v.Set(reflect.ValueOf(float32(f)))
		default:
			return fmt.Errorf("unhandled type: %v", v.Kind())
		}
		return nil
	default:
		return fmt.Errorf("unhandled token")
	}
}
