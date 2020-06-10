package json

import (
	"io"
)

const (
	ObjectStart = '{' // {
	ObjectEnd   = '}' // }
	String      = '"' // "
	Colon       = ':' // :
	Comma       = ',' // ,
	ArrayStart  = '[' // [
	ArrayEnd    = ']' // ]
	True        = 't' // t
	False       = 'f' // f
	Null        = 'n' // n
)

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
}

// NewScanner returns a new Scanner for the io.Reader r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: r,
	}
}

// Scanner implements a JSON scanner as defined in RFC 7159.
// A Scanner reads from the supplied io.Reader and produces via Next a stream
// of tokens, expressed as []byte slices.
type Scanner struct {
	stack bitvec // unused but the padding is worth up to 3% on the mb/sec
	pos   int
	r     io.Reader
	buffer
	err error
}

const optimalReadSize = 1024

func (s *Scanner) extend(elements int) int {
	oldLen := s.remaining()

	if elements == 0 || elements < optimalReadSize && s.avail() == 0 {
		// optimal read, or first read. Use optimal read size
		elements = optimalReadSize
	} else {
		// requesting a specific amount. Don't want to over-allocate the
		// buffer, limit the request to 2x current elements, or optimal
		// read size, whatever is larger.
		cap := max(optimalReadSize, oldLen*2)
		elements = min(cap, elements)
	}

	// ensure we maximize buffer use.
	elements = max(elements, s.avail())

	if s.buffer.extend(elements) == 0 {
		// could not extend
		return 0
	}

	buf := s.window()[oldLen:]
	var nread int
	nread, s.err = s.r.Read(buf)
	// give back data we did not read.
	s.releaseBack(s.remaining() - oldLen - nread)
	return nread
}

// Next returns a []byte referencing the the next lexical token in the stream.
// The []byte is valid until Next is called again.
// If the stream is at its end, or an error has occured, Next returns a zero
// length []byte slice.
//
// A valid token begins with one of the following:
//
//  { Object start
//  [ Array start
//  } Object end
//  ] Array End
//  , Literal comma
//  : Literal colon
//  t JSON true
//  f JSON false
//  n JSON null
//  " A string, possibly containing backslash escaped entites.
//  -, 0-9 A number
func (s *Scanner) Next() []byte {
	s.releaseFront(s.pos)
	w := s.window()
	pos := 0
	for {
		for _, c := range w {
			// strip any leading whitespace.
			if whitespace[c] {
				pos++
				continue
			}
			length := 0
			s.releaseFront(pos)
			switch c {
			case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
				length = 1
				s.pos = 1
			case True:
				length = s.validateToken("true")
			case False:
				length = s.validateToken("false")
			case Null:
				length = s.validateToken("null")
			case String:
				// string
				numChars := s.parseString()
				if numChars < 2 {
					return nil
				}
				length = numChars
			default:
				// ensure the number is correct.
				numChars := s.parseNumber()
				if numChars < 0 {
					return nil
				}
				length = numChars

			}
			return s.window()[:length]
		}
		// If no data is left, we need to extend
		if s.extend(0) == 0 {
			// eof
			return nil
		}
		w = s.window()[pos:]
	}
}

func (s *Scanner) validateToken(expected string) int {
	s.ensure(len(expected))
	if len(expected) > s.remaining() {
		// error, cannot be valid json.
		return 0
	}
	w := s.window()[:len(expected)]
	if string(w) != expected {
		// doesn't match
		return 0
	}
	s.pos = len(expected)
	return len(expected)
}

func (s *Scanner) parseString() int {
	pos := 1
	escaped := false
	w := s.window()[pos:]
	for {
		for _, c := range w {
			pos++
			switch {
			case c == '\\':
				escaped = true
			case escaped:
				escaped = false
			case c == '"' && !escaped:
				// finished
				s.pos = pos
				return pos
			}
		}
		// need more data from the pipe
		if s.extend(0) == 0 {
			// EOF.
			return -1
		}
		w = s.window()[pos:]
	}
}

func (s *Scanner) parseNumber() int {
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

	pos := 0
	var state uint16 = begin
	w := s.window()[pos:]
	for {
		for _, elem := range w {
			switch state {
			case begin:
				// only accept sign or digit
				if elem == '-' {
					state = sign
					break
				}
				fallthrough
			case sign:
				switch elem {
				case '0':
					state = leadingzero
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					state = anydigit1
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
				switch elem {
				case '.':
					state = decimal
				case 'e', 'E':
					state = exponent
				default:
					s.pos = pos
					return pos // finished
				}
			case decimal:
				if elem >= '0' && elem <= '9' {
					state = anydigit2
				} else {
					// error
					return -1
				}
			case anydigit2:
				switch elem {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					break
				case 'e', 'E':
					state = exponent
				default:
					s.pos = pos
					return pos // finished
				}
			case exponent:
				if elem == '+' || elem == '-' {
					state = expsign
					break
				}
				fallthrough
			case expsign:
				if elem >= '0' && elem <= '9' {
					state = anydigit3
				} else {
					// error
					return -1
				}
			case anydigit3:
				if elem >= '0' && elem <= '9' {
					break
				}
				s.pos = pos
				return pos // finished
			}
			pos++
		}

		// need more data from the pipe
		if s.extend(0) == 0 {
			// end of the item. However, not necessarily an error. Make
			// sure we are in a state that allows ending the number.
			switch state {
			case leadingzero, anydigit1, anydigit2, anydigit3:
				s.pos = pos
				return pos // finished.
			default:
				// error otherwise, the number isn't complete.
				return -1
			}
		}
		w = s.window()[pos:]
	}
}

func (s *Scanner) ensure(elems int) {
	for s.remaining() < elems {
		if s.extend(elems-s.remaining()) == 0 {
			break
		}
	}
}

// Error returns the first error encountered.
// When underlying reader is exhausted, Error returns io.EOF.
func (s *Scanner) Error() error { return s.err }
