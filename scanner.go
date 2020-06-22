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
	br  byteReader
	pos int
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
	s.br.release(s.pos)
	w := s.br.window()
	pos := 0
	for {
		for _, c := range w {
			// strip any leading whitespace.
			if whitespace[c] {
				pos++
				continue
			}
			length := 0
			s.br.release(pos)
			switch c {
			case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
				length = 1
				s.pos = 1
			case True:
				length = validateToken(&s.br, "true")
				s.pos = length
			case False:
				length = validateToken(&s.br, "false")
				s.pos = length
			case Null:
				length = validateToken(&s.br, "null")
				s.pos = length
			case String:
				// string
				length = s.parseString()
				if length < 2 {
					return nil
				}
			default:
				// ensure the number is correct.
				length = s.parseNumber()
				if length < 0 {
					return nil
				}
			}
			return s.br.window()[:length]
		}
		// If no data is left, we need to extend
		if s.br.extend() == 0 {
			// eof
			return nil
		}
		w = s.br.window()[pos:]
	}
}

func validateToken(br *byteReader, expected string) int {
	n := len(expected)
	w := br.window()
	for {
		if n <= len(w) {
			if string(w[:n]) != expected {
				// doesn't match
				return 0
			}
			return n
		}
		// If no data is left, we need to extend
		if br.extend() == 0 {
			// eof
			return 0
		}
		w = br.window()
	}
}

func (s *Scanner) parseString() int {
	pos := 1
	escaped := false
	w := s.br.window()[pos:]
	for {
		for _, c := range w {
			pos++
			switch escaped {
			case true:
				escaped = false
			case false:
				if c == '\\' {
					escaped = true
				}
				if c == '"' {
					// finished
					s.pos = pos
					return pos
				}
			}
		}
		// need more data from the pipe
		if s.br.extend() == 0 {
			// EOF.
			return -1
		}
		w = s.br.window()[pos:]
	}
}

func (s *Scanner) parseNumber() int {
	const (
		begin = iota
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
	w := s.br.window()
	// int vs uint8 costs 10% on canada.json
	var state uint8 = begin
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
		if s.br.extend() == 0 {
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
		w = s.br.window()[pos:]
	}
}

// Error returns the first error encountered.
// When underlying reader is exhausted, Error returns io.EOF.
func (s *Scanner) Error() error { return s.br.err }
