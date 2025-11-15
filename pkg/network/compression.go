package network

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// CompressionThreshold is the minimum message size (in bytes) for compression.
// Messages smaller than this are sent uncompressed.
const CompressionThreshold = 100

// CompressMessage compresses a message using zlib if it exceeds the threshold.
// Returns the compressed data and a boolean indicating if compression was applied.
func CompressMessage(data []byte) ([]byte, bool, error) {
	// Skip compression for small messages
	if len(data) < CompressionThreshold {
		return data, false, nil
	}

	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, false, fmt.Errorf("compression write failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, false, fmt.Errorf("compression close failed: %w", err)
	}

	compressed := buf.Bytes()

	// Only use compression if it actually reduces size
	// (Some data may not compress well, e.g., already encrypted data)
	if len(compressed) < len(data) {
		return compressed, true, nil
	}

	return data, false, nil
}

// DecompressMessage decompresses a zlib-compressed message.
func DecompressMessage(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return buf.Bytes(), nil
}

// EstimateCompressionRatio estimates the compression ratio for the given data.
// Returns the ratio as a percentage (e.g., 0.3 = 30% of original size).
func EstimateCompressionRatio(data []byte) float64 {
	compressed, wasCompressed, err := CompressMessage(data)
	if err != nil || !wasCompressed {
		return 1.0 // No compression
	}

	return float64(len(compressed)) / float64(len(data))
}
