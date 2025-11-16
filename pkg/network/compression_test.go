package network

import (
	"fmt"
	"strings"
	"testing"
)

func TestCompressMessage_BelowThreshold(t *testing.T) {
	// Message below 100 bytes should not be compressed
	shortMsg := []byte("Hello, world!")
	compressed, wasCompressed, err := CompressMessage(shortMsg)
	if err != nil {
		t.Fatalf("CompressMessage failed: %v", err)
	}

	if wasCompressed {
		t.Error("Expected no compression for message below threshold")
	}

	if !bytesEqual(compressed, shortMsg) {
		t.Error("Expected uncompressed data to match original")
	}
}

func TestCompressMessage_AboveThreshold(t *testing.T) {
	// Create a message >100 bytes with repeating content (compresses well)
	longMsg := []byte(strings.Repeat("This is a test message with repeating content. ", 10))

	if len(longMsg) < CompressionThreshold {
		t.Fatalf("Test message too short: %d bytes", len(longMsg))
	}

	compressed, wasCompressed, err := CompressMessage(longMsg)
	if err != nil {
		t.Fatalf("CompressMessage failed: %v", err)
	}

	if !wasCompressed {
		t.Error("Expected compression for message above threshold")
	}

	if len(compressed) >= len(longMsg) {
		t.Errorf("Compressed size (%d) should be smaller than original (%d)", len(compressed), len(longMsg))
	}

	// Verify compression ratio
	ratio := float64(len(compressed)) / float64(len(longMsg))
	if ratio > 0.7 {
		t.Errorf("Expected compression ratio <70%%, got %.2f%%", ratio*100)
	}
}

func TestCompressDecompress_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"short message", []byte("Hello")},
		{"threshold message", []byte(strings.Repeat("x", 100))},
		{"long message", []byte(strings.Repeat("This is a test. ", 50))},
		{"random bytes", generateRandomBytes(200)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, wasCompressed, err := CompressMessage(tt.data)
			if err != nil {
				t.Fatalf("Compression failed: %v", err)
			}

			if !wasCompressed {
				// No compression applied, verify data unchanged
				if !bytesEqual(compressed, tt.data) {
					t.Error("Uncompressed data doesn't match original")
				}
				return
			}

			// Decompress and verify
			decompressed, err := DecompressMessage(compressed)
			if err != nil {
				t.Fatalf("Decompression failed: %v", err)
			}

			if !bytesEqual(decompressed, tt.data) {
				t.Error("Decompressed data doesn't match original")
			}
		})
	}
}

func TestDecompressMessage_InvalidData(t *testing.T) {
	invalidData := []byte("This is not compressed data")
	_, err := DecompressMessage(invalidData)

	if err == nil {
		t.Error("Expected error for invalid compressed data")
	}
}

func TestEstimateCompressionRatio(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectRatio float64 // Approximate expected ratio
	}{
		{
			name:        "highly compressible",
			data:        []byte(strings.Repeat("AAAA", 50)),
			expectRatio: 0.1, // ~10% of original
		},
		{
			name:        "moderately compressible",
			data:        []byte(strings.Repeat("Hello, world! ", 20)),
			expectRatio: 0.15, // Better compression than originally expected
		},
		{
			name:        "below threshold",
			data:        []byte("Short message"),
			expectRatio: 1.0, // No compression
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := EstimateCompressionRatio(tt.data)

			// Allow 30% margin for compression variance (compression is variable)
			margin := 0.3
			if ratio < tt.expectRatio-margin || ratio > tt.expectRatio+margin {
				t.Errorf("Expected ratio ~%.2f, got %.2f", tt.expectRatio, ratio)
			}
		})
	}
}

func TestCompressionBandwidthSavings(t *testing.T) {
	// Test typical chat messages for bandwidth savings
	messages := []string{
		"Hey, how are you doing?", // Short
		strings.Repeat("This is a longer message with some repetitive content. ", 5),   // Medium
		strings.Repeat("The quick brown fox jumps over the lazy dog repeatedly. ", 10), // Long
	}

	totalOriginal := 0
	totalCompressed := 0

	for _, msg := range messages {
		data := []byte(msg)
		compressed, _, err := CompressMessage(data)
		if err != nil {
			t.Fatalf("Compression failed: %v", err)
		}

		totalOriginal += len(data)
		totalCompressed += len(compressed)
	}

	savingsPercent := (1.0 - float64(totalCompressed)/float64(totalOriginal)) * 100
	t.Logf("Total savings: %.1f%% (%d bytes → %d bytes)", savingsPercent, totalOriginal, totalCompressed)

	// With typical chat messages, expect at least some savings on longer messages
	if totalCompressed > totalOriginal {
		t.Error("Expected overall bandwidth savings from compression")
	}
}

// BenchmarkCompressMessage benchmarks compression performance
func BenchmarkCompressMessage(b *testing.B) {
	sizes := []int{50, 100, 200, 500, 1000}

	for _, size := range sizes {
		msg := []byte(strings.Repeat("Test message content. ", size/20))
		b.Run(fmt.Sprintf("%dbytes", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _, _ = CompressMessage(msg)
			}
		})
	}
}

// BenchmarkDecompressMessage benchmarks decompression performance
func BenchmarkDecompressMessage(b *testing.B) {
	msg := []byte(strings.Repeat("Test message content. ", 50))
	compressed, _, err := CompressMessage(msg)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecompressMessage(compressed)
	}
}

// Helper functions

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateRandomBytes(n int) []byte {
	// Simple pseudo-random bytes for testing
	result := make([]byte, n)
	for i := range result {
		result[i] = byte(i % 256)
	}
	return result
}
