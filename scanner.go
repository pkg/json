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
	br  byteReader
	pos int
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
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
	w := s.br.window(0)
loop:
	for pos, c := range w {
		// strip any leading whitespace.
		if whitespace[c] {
			continue
		}

		// simple case
		switch c {
		case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
			s.pos = pos + 1
			return w[pos:s.pos]
		}

		s.br.release(pos)
		switch c {
		case True:
			s.pos = validateToken(&s.br, "true")
		case False:
			s.pos = validateToken(&s.br, "false")
		case Null:
			s.pos = validateToken(&s.br, "null")
		case String:
			if s.parseString() < 2 {
				return nil
			}
		default:
			// ensure the number is correct.
			s.pos = s.parseNumber()
		}
		return s.br.window(0)[:s.pos]
	}

	// it's all whitespace, ignore it
	s.br.release(len(w))

	// refill buffer
	if s.br.extend() == 0 {
		// eof
		return nil
	}
	w = s.br.window(0)
	goto loop
}

func validateToken(br *byteReader, expected string) int {
	for {
		w := br.window(0)
		n := len(expected)
		if len(w) >= n {
			if string(w[:n]) != expected {
				// doesn't match
				return 0
			}
			return n
		}
		// not enough data is left, we need to extend
		if br.extend() == 0 {
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
	w := s.br.window(1)
	pos := 0
	for {
		for _, c := range w {
			pos++
			switch {
			case escaped:
				escaped = false
			case c == '"':
				// finished
				s.pos = pos + 1
				return s.pos
			case c == '\\':
				escaped = true
			}
		}
		// need more data from the pipe
		if s.br.extend() == 0 {
			// EOF.
			return 0
		}
		w = s.br.window(pos + 1)
	}
}

func (s *Scanner) parseNumber() int {
	var nd, i int
	var sawdot, sawe bool
	var buf = s.br.window(0)

	// index 0 is guarenteed to be valid
	if buf[i] == '-' {
		i++
	}

loop:
	for {
		for ; i < len(buf); i++ {
			switch c := buf[i]; true {
			case '0' <= c && c <= '9':
				nd++
				continue
			case c == '.':
				i++
				sawdot = true
			}
			break loop
		}
		// need more data from the pipe
		if s.br.extend() == 0 {
			break loop
		}
		buf = s.br.window(0)
	}
	if nd == 0 {
		return 0
	}
	if sawdot {
		nd = 0
	loop1:
		for {
			for ; i < len(buf); i++ {
				switch c := buf[i]; true {
				case '0' <= c && c <= '9':
					nd++
					continue
				case lower(c) == 'e':
					i++
					sawe = true
				}
				break loop1
			}
			// need more data from the pipe
			if s.br.extend() == 0 {
				break loop1
			}
			buf = s.br.window(0)
		}
		if nd == 0 {
			return 0
		}
	}

	if sawe {
		nd = 0
	loop2:
		for {
			for ; i < len(buf); i++ {
				switch c := buf[i]; true {
				case c == '+' || c == '-':
					continue
				case '0' <= c && c <= '9':
					nd++
					continue
				}
				break loop2
			}
			// need more data from the pipe
			if s.br.extend() == 0 {
				break loop2
			}
			buf = s.br.window(0)
		}
		if nd == 0 {
			return 0
		}
	}
	return i
}

// lower(c) is a lower-case letter if and only if
// c is either that lower-case letter or the equivalent upper-case letter.
// Instead of writing c == 'x' || c == 'X' one can write lower(c) == 'x'.
// Note that lower of non-letters can produce other non-letters.
func lower(c byte) byte {
	return c | ('x' - 'X')
}

// Error returns the first error encountered.
// When underlying reader is exhausted, Error returns io.EOF.
func (s *Scanner) Error() error { return s.br.err }
