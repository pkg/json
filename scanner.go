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

// NewScanner returns a new Scanner for the io.Reader r.
// A Scanner reads from the supplied io.Reader and produces via Next a stream
// of tokens, expressed as []byte slices.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		br: byteReader{
			r: r,
		},
	}
}

// Scanner implements a JSON scanner as defined in RFC 7159.
type Scanner struct {
	br     byteReader
	offset int
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
}

// Next returns a []byte referencing the the next lexical token in the stream.
// The []byte is valid until Next is called again.
// If the stream is at its end, or an error has occurred, Next returns a zero
// length []byte slice.
//
// A valid token begins with one of the following:
//
//	{ Object start
//	[ Array start
//	} Object end
//	] Array End
//	, Literal comma
//	: Literal colon
//	t JSON true
//	f JSON false
//	n JSON null
//	" A string, possibly containing backslash escaped entites.
//	-, 0-9 A number
func (s *Scanner) Next() []byte {
	s.br.release(s.offset)
	w := s.br.window()
	for {
		for pos, c := range w {
			// strip any leading whitespace.
			if whitespace[c] {
				continue
			}

			// simple case
			switch c {
			case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
				s.offset = pos + 1
				return w[pos:s.offset]
			}

			s.br.release(pos)
			switch c {
			case True:
				s.offset = s.validateToken("true")
			case False:
				s.offset = s.validateToken("false")
			case Null:
				s.offset = s.validateToken("null")
			case String:
				if s.parseString() < 2 {
					return nil
				}
			default:
				// ensure the number is correct.
				s.offset = s.parseNumber(c)
			}
			return s.br.window()[:s.offset]
		}

		// it's all whitespace, ignore it
		s.br.release(len(w))

		// refill buffer
		if s.br.extend() == 0 {
			// eof
			return nil
		}
		w = s.br.window()
	}
}

func (s *Scanner) validateToken(expected string) int {
	for {
		w := s.br.window()
		n := len(expected)
		if len(w) >= n {
			if string(w[:n]) != expected {
				// doesn't match
				return 0
			}
			return n
		}
		// not enough data is left, we need to extend
		if s.br.extend() == 0 {
			// eof
			return 0
		}
	}
}

// parseString returns the length of the string token
// located at the start of the window or 0 if there is no closing
// " before the end of the byteReader.
func (s *Scanner) parseString() int {
	escaped := false
	w := s.br.window()[1:]
	offset := 0
	for {
		for _, c := range w {
			offset++
			switch {
			case escaped:
				escaped = false
			case c == '"':
				// finished
				s.offset = offset + 1
				return s.offset
			case c == '\\':
				escaped = true
			}
		}
		// need more data from the pipe
		if s.br.extend() == 0 {
			// EOF.
			return 0
		}
		w = s.br.window()[offset+1:]
	}
}

func (s *Scanner) parseNumber(c byte) int {
	const (
		begin = iota
		leadingzero
		anydigit1
		decimal
		anydigit2
		exponent
		expsign
		anydigit3
	)

	offset := 0
	w := s.br.window()
	// int vs uint8 costs 10% on canada.json
	var state uint8 = begin

	// handle the case that the first character is a hyphen
	if c == '-' {
		offset++
	}

	for {
		for _, elem := range w[offset:] {
			switch state {
			case begin:
				if elem >= '1' && elem <= '9' {
					state = anydigit1
				} else if elem == '0' {
					state = leadingzero
				} else {
					// error
					return 0
				}
			case anydigit1:
				if elem >= '0' && elem <= '9' {
					// stay in this state
					break
				}
				fallthrough
			case leadingzero:
				if elem == '.' {
					state = decimal
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				return offset // finished.
			case decimal:
				if elem >= '0' && elem <= '9' {
					state = anydigit2
				} else {
					// error
					return 0
				}
			case anydigit2:
				if elem >= '0' && elem <= '9' {
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				return offset // finished.
			case exponent:
				if elem == '+' || elem == '-' {
					state = expsign
					break
				}
				fallthrough
			case expsign:
				if elem >= '0' && elem <= '9' {
					state = anydigit3
					break
				}
				// error
				return 0
			case anydigit3:
				if elem < '0' || elem > '9' {
					return offset
				}
			}
			offset++
		}

		// need more data from the pipe
		if s.br.extend() == 0 {
			// end of the item. However, not necessarily an error. Make
			// sure we are in a state that allows ending the number.
			switch state {
			case leadingzero, anydigit1, anydigit2, anydigit3:
				return offset
			default:
				// error otherwise, the number isn't complete.
				return 0
			}
		}
		w = s.br.window()
	}
}

// Error returns the first error encountered.
// When underlying reader is exhausted, Error returns io.EOF.
func (s *Scanner) Error() error { return s.br.err }
