package pool

import (
	"sync"
)

// BufferPool provides a pool of reusable byte buffers
type BufferPool struct {
	pool sync.Pool
	size int
}

// NewBufferPool creates a new buffer pool with the specified buffer size
func NewBufferPool(bufferSize int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferSize)
			},
		},
		size: bufferSize,
	}
}

// Get retrieves a buffer from the pool
func (bp *BufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf []byte) {
	// Only put back buffers of the correct size
	if cap(buf) == bp.size {
		bp.pool.Put(buf[:bp.size])
	}
}

// Size returns the buffer size for this pool
func (bp *BufferPool) Size() int {
	return bp.size
}

// Global buffer pools for common sizes
var (
	// Small buffers for headers and small messages (1KB)
	SmallBufferPool = NewBufferPool(1024)
	
	// Medium buffers for typical file chunks (64KB)
	MediumBufferPool = NewBufferPool(64 * 1024)
	
	// Large buffers for large file transfers (1MB)
	LargeBufferPool = NewBufferPool(1024 * 1024)
	
	// Extra large buffers for very large transfers (16MB)
	ExtraLargeBufferPool = NewBufferPool(16 * 1024 * 1024)
)

// GetBuffer returns a buffer of the appropriate size
func GetBuffer(size int) []byte {
	switch {
	case size <= 1024:
		return SmallBufferPool.Get()
	case size <= 64*1024:
		return MediumBufferPool.Get()
	case size <= 1024*1024:
		return LargeBufferPool.Get()
	default:
		return ExtraLargeBufferPool.Get()
	}
}

// PutBuffer returns a buffer to the appropriate pool
func PutBuffer(buf []byte) {
	size := cap(buf)
	switch {
	case size <= 1024:
		SmallBufferPool.Put(buf)
	case size <= 64*1024:
		MediumBufferPool.Put(buf)
	case size <= 1024*1024:
		LargeBufferPool.Put(buf)
	default:
		ExtraLargeBufferPool.Put(buf)
	}
}
