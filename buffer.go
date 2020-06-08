package json

// buffer implements Steven Schveighoffer's iopipe buffered window.
// A buffer window is a sliding window over a []byte slice.
//
// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
// | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | a | b | c | d | e | f |
// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//   ^       ^                       ^                   ^       ^
//   |       |                       |                   |       |
//   |       `- buffer.released      `- buffer.valid     |       |
//   |       |                       |                   |       `- buffer.buf.cap
//   |       `--- buffer.window() ---+                   |       |
//   |       `- buffer.remaining() --+                   `- buffer.buf.len
//   |                               |                           |
//   `- buffer.buf                   `- buffer.avail() ----------+
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
	return cap(b.buf) - b.remaining()
}

func (b *buffer) remaining() int {
	// return len(b.window)
	return b.valid - b.released
}

// tuning constants for buffer.extend.
const (
	newBufferSize = 8192
)

func (b *buffer) extend(request int) int {
	if b.remaining() == 0 {
		b.valid, b.released = 0, 0
	}
	if cap(b.buf)-b.valid >= request {
		// buffer has enough free space to accomodate.
		b.valid += request
		return request
	}

	if cap(b.buf)-b.remaining() >= request {
		// buffer has enough space if we move the data to the front.
		b.valid = copy(b.buf[:cap(b.buf)], b.buf[b.released:b.valid]) + request
		b.released = 0
		return request
	}

	// otherwise, we must allocate/extend a new buffer
	maxBufSize := max(cap(b.buf)*2, newBufferSize)
	request = min(request, maxBufSize-b.remaining())
	newbuf := make([]byte, max(b.remaining()+request, newBufferSize))

	b.valid = copy(newbuf, b.buf[b.released:b.valid]) + request
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
