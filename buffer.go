package json

import "io"

type buffer struct {
	buf             []byte
	released, valid int
}

func (b *buffer) releaseFront(elements int) {
	b.released += elements
}

func (b *buffer) releaseBack(elements int) {
	b.valid -= elements
}

func (b *buffer) window() []byte {
	return b.buf[b.released:b.valid]
}

func (b *buffer) avail() int {
	return len(b.buf) - (b.valid - b.released)
}

func (b *buffer) capacity() int {
	return cap(b.buf)
}

func (b *buffer) extend(request int) int {
	validElems := b.valid - b.released
	if validElems == 0 {
		b.valid, b.released = b.released, 0
	}
	if len(b.buf)-b.valid >= request {
		// buffer has enough free space to accomodate.
		b.valid += request
		return request
	}

	if len(b.buf)-validElems >= request {
		// buffer has enough space if we move the data to the front.
		copy(b.buf[0:validElems], b.buf[b.released:b.valid])
		b.released = 0
		b.valid = validElems + request
		return request
	}

	// otherwise, we must allocate/extend a new buffer
	maxBufSize := max(len(b.buf)*2, 8192)
	request = min(request, maxBufSize-validElems)
	newLen := max(validElems+request, 8192)
	newbuf := make([]byte, newLen)

	if validElems > 0 {
		copy(newbuf[0:validElems], b.buf[b.released:b.valid])
	}
	b.valid = validElems + request
	b.released = 0

	b.buf = newbuf

	return request
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type source struct {
	r   io.Reader
	buf buffer
	err error
}

func (s *source) window() []byte {
	return s.buf.window()
}

func (s *source) release(elements int) {
	s.buf.releaseFront(elements)
}

func (s *source) extend(elements int) int {
	oldLen := len(s.buf.window())
	const optimalReadSize = 8192

	if elements == 0 || elements < optimalReadSize && s.buf.capacity() == 0 {
		// optimal read, or first read. Use optimal read size
		elements = optimalReadSize
	} else {
		// requesting a specific amount. Don't want to over-allocate the
		// buffer, limit the request to 2x current elements, or optimal
		// read size, whatever is larger.
		cap := max(optimalReadSize, oldLen*2)
		if elements > cap {
			elements = cap
		}
	}

	// ensure we maximize buffer use.
	elements = max(elements, s.buf.avail())

	if s.buf.extend(elements) == 0 {
		// could not extend
		return 0
	}

	var nread int
	nread, s.err = s.r.Read(s.buf.window()[oldLen:])
	// give back data we did not read.
	s.buf.releaseBack(len(s.buf.window()) - oldLen - nread)
	return nread
}
