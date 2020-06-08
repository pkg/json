package json

import "testing"

func TestBufferExtend(t *testing.T) {
	assert := func(got, want int) {
		if got != want {
			t.Helper()
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	}

	var b buffer
	assert(b.remaining(), 0)
	assert(b.avail(), 0)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail()+b.released, cap(b.buf))

	b.extend(1)
	assert(b.remaining(), 1)
	assert(b.avail(), newBufferSize-b.remaining())
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail(), cap(b.buf))

	b.extend(1)
	assert(b.remaining(), 2)
	assert(b.avail(), newBufferSize-b.remaining())
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail(), cap(b.buf))

	b.releaseFront(1)
	assert(b.remaining(), 1)
	assert(b.avail(), newBufferSize-b.remaining())
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail(), cap(b.buf))

	b.releaseBack(1)
	assert(b.remaining(), 0)
	assert(b.avail(), newBufferSize-b.remaining())
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail(), cap(b.buf))

	b.extend(newBufferSize)
	assert(b.remaining(), newBufferSize)
	assert(b.avail(), newBufferSize-b.remaining())
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail(), cap(b.buf))

	n := b.remaining()
	n += b.extend(1)
	assert(b.remaining(), n)
	assert(b.avail(), 0)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.avail(), cap(b.buf))
}
