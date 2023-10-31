// Package json decodes JSON.
package json

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unsafe"
)

// A Decoder decodes JSON values from an input stream.
type Decoder struct {
	scanner Scanner
	state   func(*Decoder) ([]byte, error)
	stack
}

// NewDecoder returns a new Decoder for the supplied Reader r.
func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderBuffer(r, make([]byte, 8192))
}

// NewDecoderBuffer returns a new Decoder for the supplier Reader r, using
// the []byte buf provided for working storage.
func NewDecoderBuffer(r io.Reader, buf []byte) *Decoder {
	return &Decoder{
		scanner: Scanner{
			br: byteReader{
				data: buf[:0],
				r:    r,
			},
		},
		state: (*Decoder).stateValue,
	}
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

// Token returns the next JSON token in the input stream.
// At the end of the input stream, Token returns nil, io.EOF.
//
// Token guarantees that the delimiters [ ] { } it returns are
// properly nested and matched: if Token encounters an unexpected
// delimiter in the input, it will return an error.
//
// The input stream consists of basic JSON values—bool, string,
// number, and null—along with delimiters [ ] { } of type json.Delim
// to mark the start and end of arrays and objects.
// Commas and colons are elided.
//
// Note: this API is provided for compatibility with the encoding/json
// package and carries a significant allocation cost. See NextToken for
// a more efficient API.
func (d *Decoder) Token() (json.Token, error) {
	tok, err := d.NextToken()
	if err != nil {
		return nil, err
	}
	switch tok[0] {
	case '{', '[', ']', '}':
		return json.Delim(tok[0]), nil
	case 't', 'f':
		return tok[0] == 't', nil
	case 'n':
		return nil, nil
	case '"':
		return string(tok[1 : len(tok)-1]), nil
	default:
		return strconv.ParseFloat(bytesToString(tok), 64)
	}
}

// NextToken returns a []byte referencing the next logical token in the stream.
// The []byte is valid until Token is called again.
// At the end of the input stream, Token returns nil, io.EOF.
//
// Token guarantees that the delimiters [ ] { } it returns are properly nested
// and matched: if Token encounters an unexpected delimiter in the input, it
// will return an error.
//
// A valid token begins with one of the following:
//
//	{ Object start
//	[ Array start
//	} Object end
//	] Array End
//	t JSON true
//	f JSON false
//	n JSON null
//	" A string, possibly containing backslash escaped entites.
//	-, 0-9 A number
//
// Commas and colons are elided.
func (d *Decoder) NextToken() ([]byte, error) {
	return d.state(d)
}

func (d *Decoder) stateObjectString() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case '"':
		d.state = (*Decoder).stateObjectColon
		return tok, nil
	default:
		return nil, fmt.Errorf("stateObjectString: missing string key")
	}
}

func (d *Decoder) stateObjectColon() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case Colon:
		d.state = (*Decoder).stateObjectValue
		return d.NextToken()
	default:
		return tok, fmt.Errorf("stateObjectColon: expecting colon")
	}
}

func (d *Decoder) stateObjectValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*Decoder).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*Decoder).stateArrayValue
		d.push(false)
		return tok, nil
	default:
		d.state = (*Decoder).stateObjectComma
		return tok, nil
	}
}

func (d *Decoder) stateObjectComma() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case Comma:
		d.state = (*Decoder).stateObjectString
		return d.NextToken()
	default:
		return tok, fmt.Errorf("stateObjectComma: expecting comma")
	}
}

func (d *Decoder) stateArrayValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*Decoder).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*Decoder).stateArrayValue
		d.push(false)
		return tok, nil
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateArrayValue: unexpected comma")
	default:
		d.state = (*Decoder).stateArrayComma
		return tok, nil
	}
}

func (d *Decoder) stateArrayComma() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*Decoder).stateEnd
		case inObj:
			d.state = (*Decoder).stateObjectComma
		case !inObj:
			d.state = (*Decoder).stateArrayComma
		}
		return tok, nil
	case Comma:
		d.state = (*Decoder).stateArrayValue
		return d.NextToken()
	default:
		return nil, fmt.Errorf("stateArrayComma: expected comma, %v", d.stack)
	}
}

func (d *Decoder) stateValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*Decoder).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*Decoder).stateArrayValue
		d.push(false)
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateValue: unexpected comma")
	default:
		d.state = (*Decoder).stateEnd
		return tok, nil
	}
}

func (d *Decoder) stateEnd() ([]byte, error) { return nil, io.EOF }

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
	tok, err := d.NextToken()
	if err != nil {
		return err
	}
	switch tok[0] {
	case '{':
		switch v.Kind() {
		case reflect.Interface:
			if v.NumMethod() > 0 {
				return fmt.Errorf("cannot decode object into Go value of type %v", v.Type())
			}
			m, err := d.decodeMapAny()
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(m))
		case reflect.Map:
			return d.decodeMap(v)
		default:
			return fmt.Errorf("decodeValue: unhandled type: %v", v.Kind())
		}
		return nil
	case '[':
		switch v.Kind() {
		case reflect.Interface:
			if v.NumMethod() > 0 {
				return fmt.Errorf("cannot decode array into Go value of type %v", v.Type())
			}
			s, err := d.decodeSliceAny()
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(s))
		default:
			return fmt.Errorf("unhandled type: %v", v.Kind())
		}
		return nil
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
	case '"':
		switch v.Kind() {
		case reflect.Interface:
			if v.NumMethod() > 0 {
				return fmt.Errorf("cannot decode object into Go value of type %v", v.Type())
			}
			s := string(tok[1 : len(tok)-1])
			v.Set(reflect.ValueOf(s))
		case reflect.String:
			s := string(tok[1 : len(tok)-1])
			v.SetString(s)
		default:
			return fmt.Errorf("unhandled type: %v", v.Kind())
		}
		return nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		switch v.Kind() {
		case reflect.Interface:
			if v.NumMethod() > 0 {
				return fmt.Errorf("cannot decode number into Go value of type %v", v.Type())
			}
			f, err := strconv.ParseFloat(bytesToString(tok), 64)
			if err != nil {
				return fmt.Errorf("cannot convert %q to float: %v", tok, err)
			}
			v.Set(reflect.ValueOf(f))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(bytesToString(tok), 10, 64)
			if err != nil || v.OverflowInt(i) {
				return fmt.Errorf("cannot convert %q to int: %v", tok, err)
			}
			v.SetInt(i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(bytesToString(tok), 10, 64)
			if err != nil || v.OverflowUint(u) {
				return fmt.Errorf("cannot convert %q to uint: %v", tok, err)
			}
			v.SetUint(u)
		case reflect.Float64, reflect.Float32:
			f, err := strconv.ParseFloat(bytesToString(tok), v.Type().Bits())
			if err != nil || v.OverflowFloat(f) {
				return fmt.Errorf("cannot convert %q to float: %v", tok, err)
			}
			v.SetFloat(f)
		default:
			return fmt.Errorf("unhandled type: %v", v.Kind())
		}
		return nil
	default:
		return fmt.Errorf("decodeValue: unhandled token: %c", tok[0])
	}
}

func (d *Decoder) decodeValueAny() (interface{}, error) {
	tok, err := d.NextToken()
	if err != nil {
		return nil, err
	}
	switch tok[0] {
	case '{':
		return d.decodeMapAny()
	case '[':
		return d.decodeSliceAny()
	case True, False:
		return tok[0] == 't', nil
	case '"':
		return string(tok[1 : len(tok)-1]), nil
	case Null:
		return nil, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s := bytesToString(tok)
		return strconv.ParseFloat(s, 64)
	default:
		return fmt.Errorf("decodeValueAny: unhandled token: %c", tok[0]), nil
	}
}

func (d *Decoder) decodeMapAny() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for {
		tok, err := d.NextToken()
		if err != nil {
			return nil, err
		}
		if tok[0] == '}' {
			return m, nil
		}

		key := string(tok[1 : len(tok)-1])
		val, err := d.decodeValueAny()
		if err != nil {
			return nil, fmt.Errorf("decodeMapAny: %w", err)
		}
		m[key] = val
	}
}

func (d *Decoder) decodeMap(v reflect.Value) error {
	t := v.Type()
	kt := t.Key()
	if kt.Kind() != reflect.String {
		return fmt.Errorf("cannot decode object into map with key type %v", kt)
	}

	for {
		tok, err := d.NextToken()
		if err != nil {
			return err
		}
		if tok[0] == '}' {
			return nil
		}
		key := string(tok[1 : len(tok)-1])
		kv := reflect.ValueOf(key).Convert(kt)

		value := reflect.New(t.Elem()).Elem()
		if err := d.decodeValue(value); err != nil {
			return err
		}
		v.SetMapIndex(kv, value)
	}
}

func (d *Decoder) decodeSliceAny() ([]interface{}, error) {
	s := make([]interface{}, 0, 1)
	for {
		tok, err := d.NextToken()
		if err != nil {
			return nil, err
		}
		switch tok[0] {
		case ']':
			return s, nil
		case '{':
			m, err := d.decodeMapAny()
			if err != nil {
				return nil, err
			}
			s = append(s, m)
		case '[':
			sv, err := d.decodeSliceAny()
			if err != nil {
				return nil, err
			}
			s = append(s, sv)
		case True, False:
			s = append(s, tok[0] == 't')
		case '"':
			s = append(s, string(tok[1:len(tok)-1]))
		case Null:
			s = append(s, nil)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			ss := bytesToString(tok)
			f, err := strconv.ParseFloat(ss, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert %q to float: %v", tok, err)
			}
			s = append(s, f)
		}
	}
}

func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
