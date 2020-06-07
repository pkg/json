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

func (b *buffer) window() []byte {
	return b.buf[b.released:b.valid]
}

func (b *buffer) avail() int {
	return len(b.buf) - b.remaining()
}

func (b *buffer) remaining() int {
	// return len(b.window)
	return b.valid - b.released
}

func (b *buffer) extend(request int) int {
	validElems := b.remaining()
	if validElems == 0 {
		b.valid, b.released = b.released, 0
	}
	if cap(b.buf)-b.valid >= request {
		// buffer has enough free space to accomodate.
		b.valid += request
		return request
	}

	if cap(b.buf)-validElems >= request {
		// buffer has enough space if we move the data to the front.
		b.valid = copy(b.buf[0:validElems], b.buf[b.released:b.valid]) + request
		b.released = 0
		return request
	}

	// otherwise, we must allocate/extend a new buffer
	maxBufSize := max(len(b.buf)*2, 8192)
	request = min(request, maxBufSize-validElems)
	newLen := max(validElems+request, 8192)
	newbuf := make([]byte, newLen)

	b.valid = copy(newbuf[0:validElems], b.buf[b.released:b.valid]) + request
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
