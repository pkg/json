package json

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

func (b *buffer) at(pos int) byte {
	return b.buf[b.released+pos]
}

func (b *buffer) window() []byte {
	return b.buf[b.released:b.valid]
}

func (b *buffer) avail() int {
	return len(b.buf) - b.remaining()
}

func (b *buffer) remaining() int {
	return b.valid - b.released
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
