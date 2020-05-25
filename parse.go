package json

const (
	ObjectStart = '{'  // {
	ObjectEnd   = '}'  // }
	String      = '"'  // "
	Colon       = ':'  // :
	Comma       = ','  // ,
	ArrayStart  = '['  // [
	ArrayEnd    = ']'  // ]
	Number      = '0'  // - or 0-9
	True        = 't'  // t
	False       = 'f'  // f
	Null        = 'n'  // n
	EOF         = 0x00 // end of stream
	Error       = 0xff // unexpected data in stream
)

func isValue(token byte) bool {
	switch token {
	case ObjectStart, ArrayStart, String, Number, True, False, Null:
		return true
	default:
		return false
	}
}

func isWhite(c byte) bool {
	return c == ' ' || (c >= 0x09 && c <= 0x0D)
}

func jsonTok(s *source, pos *int) uint8 {
	// strip any leading whitespace. If no data is left, we need to extend
	for {
		for *pos < len(s.window()) && isWhite(s.window()[*pos]) {
			*pos++
		}

		if *pos < len(s.window()) {
			break
		}

		if s.extend(0) == 0 {
			return EOF
		}
	}

	cur := s.window()[*pos]
	switch cur {
	case ObjectStart, ObjectEnd, String, Colon, Comma, ArrayStart, ArrayEnd, True, False, Null:
		return cur
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return Number
	default:
		return Error
	}
}

type JSONItem struct {
	offset int
	length int
	token  byte
	hint   byte
}

func (i *JSONItem) data(s *source) []byte {
	return s.window()[i.offset : i.offset+i.length]
}

func ensureElems(s *source, elems int) int {
	for len(s.window()) < elems {
		if s.extend(elems-len(s.window())) == 0 {
			break
		}
	}
	return len(s.window())
}

const (
	InPlace = 1 << iota // Item is not a value, or is a string that can be used in place.
	Int                 // Item is integral (no decimal or exponent).
	Float               // number has decimal place, but no exponent
	Exp                 // number has exponent (and is float).
	Escapes             // string has escapes
)

func parseString(s *source, pos *int, hint *byte) int {
	*hint = InPlace
	// the first character must be a quote
	src := s.window()
	if len(src) == 0 || src[*pos] != '"' {
		return -1
	}
	*pos++
	origPos := *pos
	for {
		if *pos == len(src) {
			// need more data from the pipe
			if s.extend(0) == 0 {
				// EOF.
				return -1
			}
			src = s.window()
		}
		if src[*pos] == '"' {
			// finished
			*pos++
			return *pos - origPos - 1
		}
		*pos++
	}
}

func parseNumber(ss *source, pos *int, hint *uint8) int {
	const (
		begin = 1 << iota
		sign
		leadingzero
		anydigit1
		decimal
		anydigit2
		exponent
		expsign
		anydigit3
	)

	src := ss.window()
	origPos := *pos
	*hint = Int
	var s uint16 = begin
	for {
		if *pos == len(src) {
			// need more data from the pipe
			if ss.extend(0) == 0 {
				// end of the item. However, not necessarily an error. Make
				// sure we are in a state that allows ending the number.
				if s == leadingzero || s == anydigit1 || s == anydigit2 || s == anydigit3 {
					return *pos - origPos // finished.
				}
				// error otherwise, the number isn't complete.
				return -1
			}
			src = ss.window()
		}
		elem := src[*pos]
		switch s {
		case begin:
			// only accept sign or digit
			if elem == '-' {
				s = sign
				break
			}
			fallthrough
		case sign:
			switch {
			case elem == '0':
				s = leadingzero
			case elem >= '1' && elem <= '9':
				s = anydigit1
			default:
				// error
				return -1
			}
		case anydigit1:
			if elem >= '0' && elem <= '9' {
				// stay in this state
				break
			}
			fallthrough
		case leadingzero:
			if elem == '.' {
				*hint = Float
				s = decimal
			} else if elem == 'e' || elem == 'E' {
				*hint = Exp
				s = exponent
			} else {
				return *pos - origPos // finished
			}
		case decimal:
			if elem >= '0' && elem <= '9' {
				s = anydigit2
			} else {
				// error
				return -1
			}
		case anydigit2:
			if elem >= '0' && elem <= '9' {
				break
			} else if elem == 'e' || elem == 'E' {
				*hint = Exp
				s = exponent
			} else {
				return *pos - origPos // finished
			}
		case exponent:
			if elem == '+' || elem == '-' {
				s = expsign
				break
			}
			fallthrough
		case expsign:
			if elem >= '0' && elem <= '9' {
				s = anydigit3
			} else {
				// error
				return -1
			}
		case anydigit3:
			if elem >= '0' && elem <= '9' {
				break
			} else {
				return *pos - origPos // finished
			}
		}
		*pos++
	}
}

func jsonItem(s *source, pos *int) JSONItem {
	result := JSONItem{
		token:  jsonTok(s, pos),
		offset: *pos,
	}

	validateToken := func(expected string) {
		if *pos+len(expected) > len(s.window()) {
			// need to extend
			ensureElems(s, *pos+len(expected))
		}
		w := s.window()[*pos:]
		if len(expected) > len(w) {
			// error, cannot be valid json.
			result.offset = len(s.window())
			result.token = Error
			return
		}
		// can't use std.algorithm.equal here, because of autodecoding...
		for i := 0; i < len(expected); i++ {
			if w[i] != expected[i] {
				// doesn't match
				result.offset = *pos + i
				result.token = Error
				return
			}
		}
		result.length = len(expected)
		*pos += len(expected)
	}

	switch result.token {
	case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
		result.length = 1
		*pos++ // skip over the single character item
	case EOF, Error:
		// no changes to result needed.
	case True:
		validateToken("true")
	case False:
		validateToken("false")
	case Null:
		validateToken("null")
	case String:
		// string
		numChars := parseString(s, pos, &result.hint)
		if numChars < 0 {
			result.token = Error
			result.length = *pos - result.offset
		} else {
			// skip over initial quote
			result.offset++
			result.length = numChars
		}
	case Number:
		// ensure the number is correct.
		numChars := parseNumber(s, pos, &result.hint)
		if numChars < 0 {
			result.token = Error
			result.length = *pos - result.offset
		} else {
			result.length = numChars
		}
	}
	return result
}

const (
	StateBegin  = 1 << iota // next item should be either an Object or Array
	StateFirst              // Just started a new object or array.
	StateMember             // Expect next member (name for object, value for array_
	StateColon              // Expect colon (Object only)
	StateValue              // Expect value
	StateComma              // Expect comma or end of collection.
	StateEnd                // there shouldn't be any more items
)

type Tokenizer struct {
	source   *source
	state    uint8
	stack    uint64
	stackLen int
	pos      int
}

func (t *Tokenizer) inObj() bool {
	if t.stackLen == 0 {
		return false
	}
	return t.stack&(1<<t.stackLen) != 0
}

func (t *Tokenizer) pushContainer(isObj bool) {
	t.stack <<= 1
	if t.inObj() {
		t.stack |= 0x1
	}
	t.stackLen++
}

func (t *Tokenizer) popContainer() {
	t.stackLen--
	if t.stackLen == 0 {
		t.state = StateEnd
	} else {
		t.state = StateComma
	}
}

func (t *Tokenizer) Next() JSONItem {
	if t.state == StateEnd {
		// return an EOF item, even if the stream is not done.
		return JSONItem{
			offset: t.pos,
			length: 0,
			token:  EOF,
		}
	}
	item := jsonItem(t.source, &t.pos)
	switch t.state {
	case StateBegin:
		// item needs to be an ObjectStart or ArrayStart
		if item.token == ObjectStart || item.token == ArrayStart {
			t.state = StateFirst
			t.pushContainer(item.token == ObjectStart)
		} else {
			item.token = Error
		}
	case StateFirst:
		// allow ending of the container
		tok := byte(ArrayEnd)
		if t.inObj() {
			tok = ObjectEnd
		}
		if item.token == tok {
			t.popContainer()
			break
		}
		fallthrough
	case StateMember:
		if t.inObj() {
			if item.token == String {
				t.state = StateColon
			} else {
				item.token = Error
			}
			break
		}
		fallthrough
	case StateValue:
		if isValue(item.token) {
			if item.token == ObjectStart || item.token == ArrayStart {
				t.pushContainer(item.token == ObjectStart)
				t.state = StateFirst
			} else {
				t.state = StateComma
			}
		} else {
			item.token = Error
		}
	case StateColon:
		// requires colon
		if item.token == Colon {
			t.state = StateValue
		} else {
			item.token = Error
		}

	case StateComma:
		// can end the object here, or get a comma
		tok := byte(ArrayEnd)
		if t.inObj() {
			tok = ObjectEnd
		}
		if item.token == tok {
			t.popContainer()
		} else if item.token == Comma {
			t.state = StateMember
		} else {
			item.token = Error
		}
	case StateEnd:
		// this is handled outside the switch statement
		panic("error")
	}
	return item
}
