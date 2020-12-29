package json

import "io"

// A byteReader implements a sliding window over an io.Reader.
type byteReader struct {
	data   []byte
	offset int
	r      io.Reader
	err    error
}

// release discards n bytes from the front of the window.
func (b *byteReader) release(n int) {
	b.offset += n
}

// window returns the current window.
// The window is invalidated by calls to release or extend.
func (b *byteReader) window(offset int) []byte {
	return b.data[b.offset+offset:]
}

// tuning constants for byteReader.extend.
const (
	newBufferSize = 4096
	minReadSize   = newBufferSize >> 2
)

// extend extends the window with data from the underlying reader.
func (b *byteReader) extend() int {
	if b.err != nil {
		return 0
	}

	remaining := len(b.data) - b.offset
	if remaining == 0 {
		b.data = b.data[:0]
		b.offset = 0
	}
	if cap(b.data)-len(b.data) >= minReadSize {
		// nothing to do, enough space exists between len and cap.
	} else if cap(b.data)-remaining >= minReadSize {
		// buffer has enough space if we move the data to the front.
		b.compact()
	} else {
		// otherwise, we must allocate/extend a new buffer
		b.grow()
	}
	remaining += b.offset
	n, err := b.r.Read(b.data[remaining:cap(b.data)])
	// reduce length to the existing plus the data we read.
	b.data = b.data[:remaining+n]
	b.err = err
	return n
}

// grow grows the buffer, moving the active data to the front.
func (b *byteReader) grow() {
	buf := make([]byte, max(cap(b.data)*2, newBufferSize))
	copy(buf, b.data[b.offset:])
	b.data = buf
	b.offset = 0
}

// compact moves the active data to the front of the buffer.
func (b *byteReader) compact() {
	copy(b.data, b.data[b.offset:])
	b.offset = 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
