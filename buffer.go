package json

// buffer implements Steven Schveighoffer's iopipe buffered window.
// A buffer window is a sliding window over a []byte slice.
//
// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
// | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | a | b | c | d | e | f |
// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//   ^       ^                       ^                           ^
//   |       |                       |                           |
//   |       `- buffer.released      `- buffer.buf.len           |
//   |       |                       |                           `- buffer.buf.cap
//   |       `--- buffer.window() ---+                           |
//   |       `- buffer.remaining() --+                           |
//   |                               |                           |
//   `- buffer.buf                   `- buffer.avail() ----------+
type buffer struct {
	buf      []byte
	released int
}

func (b *buffer) releaseFront(elements int) {
	b.released += elements
}

func (b *buffer) releaseBack(elements int) {
	b.buf = b.buf[:len(b.buf)-elements]
}

func (b *buffer) window() []byte {
	return b.buf[b.released:]
}

func (b *buffer) avail() int {
	return cap(b.buf) - b.remaining()
}

func (b *buffer) remaining() int {
	// return len(b.window)
	return len(b.buf) - b.released
}

// tuning constants for buffer.extend.
const (
	newBufferSize = 8192
)

func (b *buffer) extend(request int) int {
	if b.remaining() == 0 {
		b.buf, b.released = b.buf[:0], 0
	}
	if cap(b.buf)-len(b.buf) >= request {
		// space exists between len and cap, extend the slice len
		// towards cap.
		b.buf = b.buf[:len(b.buf)+request]
		return request
	}

	if cap(b.buf)-b.remaining() >= request {
		// buffer has enough space if we move the data to the front.
		copy(b.buf[:b.remaining()+request], b.buf[b.released:])
		b.released = 0
		return request
	}

	// otherwise, we must allocate/extend a new buffer
	maxBufSize := max(cap(b.buf)*2, newBufferSize)
	request = min(request, maxBufSize-b.remaining())
	newlen := max(b.remaining()+request, newBufferSize)
	newbuf := make([]byte, newlen)

	n := copy(newbuf, b.buf[b.released:])
	b.buf = newbuf[:n+request]

	b.released = 0
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
