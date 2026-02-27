package network

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

// CompressionThreshold is the minimum message size (in bytes) for compression.
// Messages smaller than this are sent uncompressed.
const CompressionThreshold = 100

// MaxDecompressedSize is the maximum allowed size for decompressed data.
// This protects against decompression bomb attacks where a small compressed
// payload expands to an extremely large size. Set to 10MB to accommodate
// large game state snapshots while preventing memory exhaustion.
const MaxDecompressedSize = 10 * 1024 * 1024 // 10 MB

// ErrDecompressedSizeExceeded indicates the decompressed data exceeds MaxDecompressedSize.
var ErrDecompressedSizeExceeded = errors.New("decompressed data exceeds maximum allowed size")

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
		logrus.WithFields(logrus.Fields{
			"bytes_original": len(data),
			"error":          err.Error(),
		}).Error("compression write failed")
		return nil, false, fmt.Errorf("compression write failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"bytes_original": len(data),
			"error":          err.Error(),
		}).Error("compression close failed")
		return nil, false, fmt.Errorf("compression close failed: %w", err)
	}

	compressed := buf.Bytes()

	// Only use compression if it actually reduces size
	// (Some data may not compress well, e.g., already encrypted data)
	if len(compressed) < len(data) {
		logrus.WithFields(logrus.Fields{
			"bytes_original":   len(data),
			"bytes_compressed": len(compressed),
			"ratio":            float64(len(compressed)) / float64(len(data)),
		}).Debug("message compressed")
		return compressed, true, nil
	}

	return data, false, nil
}

// DecompressMessage decompresses a zlib-compressed message.
// Returns ErrDecompressedSizeExceeded if the decompressed data exceeds MaxDecompressedSize.
func DecompressMessage(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"bytes_compressed": len(data),
			"error":            err.Error(),
		}).Error("failed to create zlib reader")
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	// Use LimitReader to protect against decompression bombs.
	// Read up to MaxDecompressedSize + 1 to detect if limit is exceeded.
	limitedReader := io.LimitReader(reader, MaxDecompressedSize+1)

	var buf bytes.Buffer
	n, err := io.Copy(&buf, limitedReader)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"bytes_compressed": len(data),
			"error":            err.Error(),
		}).Error("decompression failed")
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	// If we read exactly MaxDecompressedSize+1, the data exceeded the limit
	if n > MaxDecompressedSize {
		logrus.WithFields(logrus.Fields{
			"bytes_compressed":   len(data),
			"bytes_decompressed": n,
			"max_size":           MaxDecompressedSize,
		}).Warn("decompressed data exceeds maximum allowed size")
		return nil, ErrDecompressedSizeExceeded
	}

	logrus.WithFields(logrus.Fields{
		"bytes_compressed":   len(data),
		"bytes_decompressed": n,
		"ratio":              float64(len(data)) / float64(n),
	}).Debug("message decompressed")

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
