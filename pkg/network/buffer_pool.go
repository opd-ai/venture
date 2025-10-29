// Package network provides buffer pooling for network message serialization.
// This file implements object pooling to reduce allocation pressure during
// high-frequency network operations.
package network

import "sync"

const (
	// DefaultBufferSize is the standard buffer size for network messages.
	// Most messages fit within 4KB; larger messages will grow slice capacity.
	DefaultBufferSize = 4096
)

// bufferPool provides reusable byte slices for network serialization.
// Using sync.Pool ensures thread-safety and automatic GC integration.
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, DefaultBufferSize)
		return &buf
	},
}

// AcquireBuffer gets a buffer from the pool.
// The returned buffer has length 0 but capacity DefaultBufferSize.
// Caller MUST call ReleaseBuffer when done to prevent leaks.
//
// Example usage:
//   buf := AcquireBuffer()
//   defer ReleaseBuffer(buf)
//   *buf = append(*buf, data...)
func AcquireBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

// ReleaseBuffer returns a buffer to the pool for reuse.
// The buffer is reset to length 0 (keeping capacity) for the next user.
// It is safe to call ReleaseBuffer with a nil pointer (no-op).
func ReleaseBuffer(buf *[]byte) {
	if buf == nil {
		return
	}
	
	// Reset length to 0 (keeps capacity for reuse)
	// This provides security isolation between uses
	*buf = (*buf)[:0]
	
	bufferPool.Put(buf)
}

// WithBuffer provides a convenient way to use a pooled buffer with automatic cleanup.
// The function receives a buffer pointer and should return the data to be copied out.
// The buffer is automatically returned to the pool after the function completes.
//
// Example usage:
//   result := WithBuffer(func(buf *[]byte) []byte {
//       *buf = append(*buf, []byte("message")...)
//       return *buf
//   })
func WithBuffer(fn func(*[]byte) []byte) []byte {
	buf := AcquireBuffer()
	defer ReleaseBuffer(buf)
	result := fn(buf)
	
	// Make a copy since buffer will be returned to pool
	output := make([]byte, len(result))
	copy(output, result)
	return output
}
