// Package engine provides component serialization helpers.
// This file implements binary serialization utilities for network transmission.
package engine

import (
	"encoding/binary"
	"errors"
	"math"
)

// ErrInvalidComponentData is returned when deserialization encounters invalid data
var ErrInvalidComponentData = errors.New("invalid component data")

// writeFloat64 writes a float64 to the buffer in little-endian format.
func writeFloat64(buf []byte, v float64) {
	binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
}

// readFloat64 reads a float64 from the buffer in little-endian format.
func readFloat64(buf []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(buf))
}

// writeInt32 writes an int32 to the buffer in little-endian format.
func writeInt32(buf []byte, v int32) {
	binary.LittleEndian.PutUint32(buf, uint32(v))
}

// readInt32 reads an int32 from the buffer in little-endian format.
func readInt32(buf []byte) int32 {
	return int32(binary.LittleEndian.Uint32(buf))
}

// writeUint64 writes a uint64 to the buffer in little-endian format.
func writeUint64(buf []byte, v uint64) {
	binary.LittleEndian.PutUint64(buf, v)
}

// readUint64 reads a uint64 from the buffer in little-endian format.
func readUint64(buf []byte) uint64 {
	return binary.LittleEndian.Uint64(buf)
}

// writeString writes a length-prefixed string to the buffer.
// Format: length(4 bytes) + data
func writeString(buf []byte, s string) int {
	writeInt32(buf, int32(len(s)))
	copy(buf[4:], []byte(s))
	return 4 + len(s)
}

// readString reads a length-prefixed string from the buffer.
// Returns the string and the number of bytes consumed.
// Returns an error via empty string if buffer is too small or length is invalid.
func readString(buf []byte) (string, int) {
	if len(buf) < 4 {
		return "", 0
	}
	length := int(readInt32(buf))
	if length <= 0 {
		return "", 4
	}
	if len(buf) < 4+length {
		return "", 4
	}
	return string(buf[4 : 4+length]), 4 + length
}

// writeBool writes a boolean to the buffer (1 byte: 0 or 1).
func writeBool(buf []byte, v bool) {
	if v {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
}

// readBool reads a boolean from the buffer.
func readBool(buf []byte) bool {
	return buf[0] != 0
}
