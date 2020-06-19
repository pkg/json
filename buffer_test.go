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
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	b.extend(1)
	assert(b.remaining(), 1)
	assert(b.avail(), newBufferSize-b.remaining()-b.released)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	b.extend(1)
	assert(b.remaining(), newBufferSize)
	assert(b.avail(), newBufferSize-b.remaining()-b.released)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	b.releaseFront(1)
	assert(b.remaining(), newBufferSize-1)
	assert(b.avail(), newBufferSize-b.remaining()-b.released)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	b.releaseBack(1)
	assert(b.remaining(), newBufferSize-2)
	assert(b.avail(), newBufferSize-b.remaining()-b.released)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	b.extend(newBufferSize)
	assert(b.remaining(), newBufferSize*2-2)
	assert(b.avail(), b.released)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	n := b.remaining()
	n += b.extend(1)
	assert(b.remaining(), n)
	assert(b.avail(), 0)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	s := b.remaining()
	b.releaseBack(newBufferSize / 4)
	assert(b.remaining(), s-(newBufferSize/4))
	assert(b.avail(), newBufferSize/4)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))

	n = b.remaining()
	n += b.extend(newBufferSize)
	assert(b.remaining(), n)
	assert(b.avail(), 0)
	assert(len(b.window()), b.remaining())
	assert(b.remaining()+b.released+b.avail(), cap(b.buf))
}
