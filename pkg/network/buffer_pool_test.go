package network

import (
	"testing"
	"unsafe"
)

func TestAcquireBuffer_ReturnsBuffer(t *testing.T) {
	buf := AcquireBuffer()
	if buf == nil {
		t.Fatal("AcquireBuffer returned nil")
	}
	if cap(*buf) != DefaultBufferSize {
		t.Errorf("Buffer capacity = %d, want %d", cap(*buf), DefaultBufferSize)
	}
	if len(*buf) != 0 {
		t.Errorf("Buffer length = %d, want 0", len(*buf))
	}
	ReleaseBuffer(buf)
}

func TestReleaseBuffer_ResetsLength(t *testing.T) {
	buf := AcquireBuffer()
	*buf = append(*buf, []byte{1, 2, 3, 4, 5}...)

	if len(*buf) != 5 {
		t.Fatalf("Buffer length before release = %d, want 5", len(*buf))
	}

	ReleaseBuffer(buf)

	buf2 := AcquireBuffer()
	if len(*buf2) != 0 {
		t.Errorf("Buffer not reset: length = %d, want 0", len(*buf2))
	}
	ReleaseBuffer(buf2)
}

func TestReleaseBuffer_NilSafe(t *testing.T) {
	// Should not panic with nil buffer
	ReleaseBuffer(nil)
}

func TestWithBuffer_AutomaticCleanup(t *testing.T) {
	result := WithBuffer(func(buf *[]byte) []byte {
		*buf = append(*buf, []byte("test data")...)
		return *buf
	})

	if string(result) != "test data" {
		t.Errorf("Result = %q, want %q", result, "test data")
	}
}

func TestWithBuffer_ReturnsCopy(t *testing.T) {
	var bufAddr uintptr

	result := WithBuffer(func(buf *[]byte) []byte {
		*buf = append(*buf, []byte("data")...)
		bufAddr = uintptr(unsafe.Pointer(&(*buf)[0]))
		return *buf
	})

	resultAddr := uintptr(unsafe.Pointer(&result[0]))

	// Result should be a different slice (copy)
	if bufAddr == resultAddr {
		t.Error("WithBuffer did not return a copy of the data")
	}
}

func TestBufferPool_Reuse(t *testing.T) {
	// First allocation
	buf1 := AcquireBuffer()
	addr1 := uintptr(unsafe.Pointer(buf1))
	*buf1 = append(*buf1, []byte("test")...)
	ReleaseBuffer(buf1)

	// Second allocation should reuse same buffer
	buf2 := AcquireBuffer()
	addr2 := uintptr(unsafe.Pointer(buf2))

	if addr1 != addr2 {
		t.Logf("Different buffer addresses (pool may have allocated new): addr1=%v, addr2=%v", addr1, addr2)
		// Note: This is not necessarily an error - sync.Pool may allocate new buffers
		// We can't guarantee reuse, but we can verify the buffer was cleared
	}

	if len(*buf2) != 0 {
		t.Errorf("Buffer not cleared after reuse: length = %d, want 0", len(*buf2))
	}

	ReleaseBuffer(buf2)
}

func TestBufferPool_NoMemoryLeaks(t *testing.T) {
	// Create and release many buffers
	// If there's a leak, test will consume excessive memory
	for i := 0; i < 10000; i++ {
		buf := AcquireBuffer()
		*buf = append(*buf, make([]byte, 1024)...) // 1KB per buffer
		ReleaseBuffer(buf)
	}

	// If test completes without OOM, pooling is working
}

func BenchmarkBufferPooling(b *testing.B) {
	b.ReportAllocs()

	b.Run("WithPooling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := AcquireBuffer()
			*buf = append(*buf, []byte("message payload data for network transmission")...)
			ReleaseBuffer(buf)
		}
	})

	b.Run("WithoutPooling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 0, DefaultBufferSize)
			buf = append(buf, []byte("message payload data for network transmission")...)
			_ = buf
		}
	})
}

func BenchmarkWithBuffer(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := WithBuffer(func(buf *[]byte) []byte {
			*buf = append(*buf, []byte("test message")...)
			return *buf
		})
		_ = result
	}
}

// BenchmarkBufferPool_ConcurrentAccess tests thread safety
func BenchmarkBufferPool_ConcurrentAccess(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := AcquireBuffer()
			*buf = append(*buf, []byte("concurrent data")...)
			ReleaseBuffer(buf)
		}
	})
}
